// Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package botanist

import (
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	kubernetesfake "github.com/gardener/gardener/pkg/client/kubernetes/fake"
	mockdnsrecord "github.com/gardener/gardener/pkg/component/extensions/dnsrecord/mock"
	"github.com/gardener/gardener/pkg/features"
	"github.com/gardener/gardener/pkg/gardenlet/apis/config"
	gardenletfeatures "github.com/gardener/gardener/pkg/gardenlet/features"
	"github.com/gardener/gardener/pkg/operation"
	"github.com/gardener/gardener/pkg/operation/garden"
	shootpkg "github.com/gardener/gardener/pkg/operation/shoot"
	gardenerutils "github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/gardener/gardener/pkg/utils/test"
)

var _ = Describe("dns", func() {
	const (
		seedNS  = "test-ns"
		shootNS = "shoot-ns"
	)

	var (
		b                        *Botanist
		seedClient, gardenClient client.Client
		s                        *runtime.Scheme

		dnsEntryTTL int64 = 1234
	)

	BeforeEach(func() {
		b = &Botanist{
			Operation: &operation.Operation{
				Config: &config.GardenletConfiguration{
					Controllers: &config.GardenletControllerConfiguration{
						Shoot: &config.ShootControllerConfiguration{
							DNSEntryTTLSeconds: &dnsEntryTTL,
						},
					},
				},
				Shoot: &shootpkg.Shoot{
					Components: &shootpkg.Components{
						Extensions: &shootpkg.Extensions{},
					},
					SeedNamespace: seedNS,
				},
				Garden: &garden.Garden{},
				Logger: logr.Discard(),
			},
		}
		b.Shoot.SetInfo(&gardencorev1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{Namespace: shootNS},
		})

		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).NotTo(HaveOccurred())

		gardenClient = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
		seedClient = fake.NewClientBuilder().WithScheme(s).Build()

		renderer := chartrenderer.NewWithServerVersion(&version.Info{})
		chartApplier := kubernetes.NewChartApplier(renderer, kubernetes.NewApplier(seedClient, meta.NewDefaultRESTMapper([]schema.GroupVersion{})))
		Expect(chartApplier).NotTo(BeNil(), "should return chart applier")

		b.GardenClient = gardenClient
		b.SeedClientSet = kubernetesfake.NewClientSetBuilder().
			WithClient(seedClient).
			WithChartApplier(chartApplier).
			Build()
	})

	Context("NeedsExternalDNS", func() {
		It("should be false when Shoot's DNS is nil", func() {
			b.Shoot.GetInfo().Spec.DNS = nil
			Expect(b.NeedsExternalDNS()).To(BeFalse())
		})

		It("should be false when Shoot DNS's domain is nil", func() {
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: nil}
			Expect(b.NeedsExternalDNS()).To(BeFalse())
		})

		It("should be false when Shoot ExternalClusterDomain is nil", func() {
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
			b.Shoot.ExternalClusterDomain = nil
			Expect(b.NeedsExternalDNS()).To(BeFalse())
		})

		It("should be false when Shoot ExternalClusterDomain is in nip.io", func() {
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
			b.Shoot.ExternalClusterDomain = pointer.String("foo.nip.io")
			Expect(b.NeedsExternalDNS()).To(BeFalse())
		})

		It("should be false when Shoot ExternalDomain is nil", func() {
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
			b.Shoot.ExternalClusterDomain = pointer.String("baz")
			b.Shoot.ExternalDomain = nil

			Expect(b.NeedsExternalDNS()).To(BeFalse())
		})

		It("should be false when Shoot ExternalDomain provider is unamanaged", func() {
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
			b.Shoot.ExternalClusterDomain = pointer.String("baz")
			b.Shoot.ExternalDomain = &gardenerutils.Domain{Provider: "unmanaged"}

			Expect(b.NeedsExternalDNS()).To(BeFalse())
		})

		It("should be true when Shoot ExternalDomain provider is valid", func() {
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
			b.Shoot.ExternalClusterDomain = pointer.String("baz")
			b.Shoot.ExternalDomain = &gardenerutils.Domain{Provider: "valid-provider"}

			Expect(b.NeedsExternalDNS()).To(BeTrue())
		})
	})

	Context("NeedsInternalDNS", func() {
		It("should be false when the internal domain is nil", func() {
			b.Garden.InternalDomain = nil
			Expect(b.NeedsInternalDNS()).To(BeFalse())
		})

		It("should be false when the internal domain provider is unmanaged", func() {
			b.Garden.InternalDomain = &gardenerutils.Domain{Provider: "unmanaged"}
			Expect(b.NeedsInternalDNS()).To(BeFalse())
		})

		It("should be true when the internal domain provider is not unmanaged", func() {
			b.Garden.InternalDomain = &gardenerutils.Domain{Provider: "some-provider"}
			Expect(b.NeedsInternalDNS()).To(BeTrue())
		})
	})

	Context("APIServerSNIEnabled", func() {
		BeforeEach(func() {
			gardenletfeatures.RegisterFeatureGates()
		})

		It("returns true when feature gate is enabled", func() {
			DeferCleanup(test.WithFeatureGate(features.DefaultFeatureGate, features.APIServerSNI, true))
			b.Garden.InternalDomain = &gardenerutils.Domain{Provider: "some-provider"}
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
			b.Shoot.ExternalClusterDomain = pointer.String("baz")
			b.Shoot.ExternalDomain = &gardenerutils.Domain{Provider: "valid-provider"}

			Expect(b.APIServerSNIEnabled()).To(BeTrue())
		})
	})

	Context("newDNSComponentsTargetingAPIServerAddress", func() {
		var (
			ctrl              *gomock.Controller
			externalDNSRecord *mockdnsrecord.MockInterface
			internalDNSRecord *mockdnsrecord.MockInterface
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			externalDNSRecord = mockdnsrecord.NewMockInterface(ctrl)
			internalDNSRecord = mockdnsrecord.NewMockInterface(ctrl)

			b.APIServerAddress = "1.2.3.4"
			b.Shoot.Components.Extensions.ExternalDNSRecord = externalDNSRecord
			b.Shoot.Components.Extensions.InternalDNSRecord = internalDNSRecord
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("sets internal and external DNSRecords", func() {
			b.Shoot.GetInfo().Status.ClusterIdentity = pointer.String("shoot-cluster-identity")
			b.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
			b.Shoot.InternalClusterDomain = "bar"
			b.Shoot.ExternalClusterDomain = pointer.String("baz")
			b.Shoot.ExternalDomain = &gardenerutils.Domain{Provider: "valid-provider"}
			b.Garden.InternalDomain = &gardenerutils.Domain{Provider: "valid-provider"}

			externalDNSRecord.EXPECT().SetRecordType(extensionsv1alpha1.DNSRecordTypeA)
			externalDNSRecord.EXPECT().SetValues([]string{"1.2.3.4"})
			internalDNSRecord.EXPECT().SetRecordType(extensionsv1alpha1.DNSRecordTypeA)
			internalDNSRecord.EXPECT().SetValues([]string{"1.2.3.4"})

			b.newDNSComponentsTargetingAPIServerAddress()
		})
	})
})
