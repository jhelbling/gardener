// Copyright 2023 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package tokenrequest_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	. "github.com/gardener/gardener/pkg/utils/gardener/tokenrequest"
	secretsutils "github.com/gardener/gardener/pkg/utils/secrets"
	secretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager"
	fakesecretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager/fake"
)

var _ = Describe("Secrets", func() {
	var (
		ctx = context.TODO()

		namespace          = "foo-bar"
		c                  client.Client
		fakeSecretsManager secretsmanager.Interface
	)

	BeforeEach(func() {
		c = fakeclient.NewClientBuilder().Build()
		fakeSecretsManager = fakesecretsmanager.New(c, namespace)

		_, err := fakeSecretsManager.Generate(
			ctx,
			&secretsutils.CertificateSecretConfig{Name: v1beta1constants.SecretNameCACluster, CommonName: "kubernetes", CertType: secretsutils.CACert},
		)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("GenerateGenericTokenKubeconfig", func() {
		It("should generate the generic token kubeconfig", func() {
			secret, err := GenerateGenericTokenKubeconfig(ctx, fakeSecretsManager, namespace, "kube-apiserver")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(secret), secret)).To(Succeed())
		})
	})

	Describe("#RenewAccessSecrets", func() {
		It("should remove the renew-timestamp annotation from all access secrets", func() {
			var (
				secret1 = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "secret1",
						Namespace:   namespace,
						Annotations: map[string]string{"serviceaccount.resources.gardener.cloud/token-renew-timestamp": "foo"},
					},
				}
				secret2 = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "secret2",
						Namespace:   namespace,
						Annotations: map[string]string{"serviceaccount.resources.gardener.cloud/token-renew-timestamp": "foo"},
						Labels:      map[string]string{"resources.gardener.cloud/purpose": "token-requestor"},
					},
				}
				secret3 = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "secret3",
						Namespace:   namespace,
						Annotations: map[string]string{"serviceaccount.resources.gardener.cloud/token-renew-timestamp": "foo"},
						Labels:      map[string]string{"resources.gardener.cloud/purpose": "token-requestor"},
					},
				}
			)

			Expect(c.Create(ctx, secret1)).To(Succeed())
			Expect(c.Create(ctx, secret2)).To(Succeed())
			Expect(c.Create(ctx, secret3)).To(Succeed())

			Expect(RenewAccessSecrets(ctx, c, namespace)).To(Succeed())

			Expect(c.Get(ctx, client.ObjectKeyFromObject(secret1), secret1)).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKeyFromObject(secret2), secret2)).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKeyFromObject(secret3), secret3)).To(Succeed())

			Expect(secret1.Annotations).To(HaveKey("serviceaccount.resources.gardener.cloud/token-renew-timestamp"))
			Expect(secret2.Annotations).NotTo(HaveKey("serviceaccount.resources.gardener.cloud/token-renew-timestamp"))
			Expect(secret3.Annotations).NotTo(HaveKey("serviceaccount.resources.gardener.cloud/token-renew-timestamp"))
		})
	})

	Describe("#IsTokenPopulated", func() {
		var (
			kubeconfigWithToken = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: AAAA
    server: https://foobar
  name: garden
contexts:
- context:
    cluster: garden
    user: garden
  name: garden
current-context: garden
kind: Config
preferences: {}
users:
- name: garden
  user:
    token: bar
`
			kubeconfigWithoutToken = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: AAAA
    server: https://foobar
  name: garden
contexts:
- context:
    cluster: garden
    user: garden
  name: garden
current-context: garden
kind: Config
preferences: {}
users:
- name: garden
  user:
    token: ""
`
			kubeconfigWrongContext = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: AAAA
    server: https://foobar
  name: garden
contexts:
- context:
    cluster: garden
    user: garden
  name: garden
current-context: foo
kind: Config
preferences: {}
users:
- name: garden
  user:
    token: bar
`
			kubeconfigNoUser = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: AAAA
    server: https://foobar
  name: garden
contexts:
- context:
    cluster: garden
    user: garden
  name: garden
- context:
  cluster: garden
  user: foo
name: foo
current-context: foo
kind: Config
preferences: {}
users:
- name: garden
  user:
    token: bar
`
		)
		DescribeTable("#IsTokenPopulated",
			func(kubeconfig string, result bool) {
				secret := corev1.Secret{
					Data: map[string][]byte{"kubeconfig": []byte(kubeconfig)},
				}
				populated, err := IsTokenPopulated(&secret)
				Expect(err).To(Succeed())
				Expect(populated).To(Equal(result))
			},
			Entry("kubeconfig with token should return true", kubeconfigWithToken, true),
			Entry("kubeconfig without token should return false", kubeconfigWithoutToken, false),
			Entry("kubeconfig with wrong context should return false", kubeconfigWrongContext, false),
			Entry("kubeconfig without user should return false", kubeconfigNoUser, false),
		)
	})
})
