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

package kubeapiserverexposure

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/component"
	kubeapiserverconstants "github.com/gardener/gardener/pkg/component/kubeapiserver/constants"
	"github.com/gardener/gardener/pkg/controllerutils"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
)

var (
	//go:embed templates/envoyfilter.yaml
	envoyFilterSpecTemplateContent string
	envoyFilterSpecTemplate        *template.Template
)

func init() {
	envoyFilterSpecTemplate = template.Must(template.
		New("envoy-filter-spec").
		Funcs(sprig.TxtFuncMap()).
		Parse(envoyFilterSpecTemplateContent),
	)
}

// SNIValues configure the kube-apiserver service SNI.
type SNIValues struct {
	Hosts               []string
	APIServerProxy      *APIServerProxy
	IstioIngressGateway IstioIngressGateway
}

// APIServerProxy contains values for the APIServer proxy protocol configuration.
type APIServerProxy struct {
	NamespaceUID       types.UID
	APIServerClusterIP string
}

// IstioIngressGateway contains the values for istio ingress gateway configuration.
type IstioIngressGateway struct {
	Namespace string
	Labels    map[string]string
}

// NewSNI creates a new instance of DeployWaiter which deploys Istio resources for
// kube-apiserver SNI access.
func NewSNI(
	client client.Client,
	applier kubernetes.Applier,
	name string,
	namespace string,
	valuesFunc func() *SNIValues,
) component.DeployWaiter {
	if valuesFunc == nil {
		valuesFunc = func() *SNIValues { return &SNIValues{} }
	}

	return &sni{
		client:     client,
		applier:    applier,
		name:       name,
		namespace:  namespace,
		valuesFunc: valuesFunc,
	}
}

type sni struct {
	client     client.Client
	applier    kubernetes.Applier
	name       string
	namespace  string
	valuesFunc func() *SNIValues
}

type envoyFilterTemplateValues struct {
	*APIServerProxy
	IngressGatewayLabels map[string]string
	Name                 string
	Namespace            string
	Host                 string
	Port                 int
}

func (s *sni) Deploy(ctx context.Context) error {
	var (
		values = s.valuesFunc()

		destinationRule = s.emptyDestinationRule()
		gateway         = s.emptyGateway()
		virtualService  = s.emptyVirtualService()

		hostName        = fmt.Sprintf("%s.%s.svc.%s", s.name, s.namespace, gardencorev1beta1.DefaultDomain)
		envoyFilterSpec bytes.Buffer
	)

	if values.APIServerProxy != nil {
		envoyFilter := s.emptyEnvoyFilter()

		if err := envoyFilterSpecTemplate.Execute(&envoyFilterSpec, envoyFilterTemplateValues{
			APIServerProxy:       values.APIServerProxy,
			IngressGatewayLabels: values.IstioIngressGateway.Labels,
			Name:                 envoyFilter.Name,
			Namespace:            envoyFilter.Namespace,
			Host:                 hostName,
			Port:                 kubeapiserverconstants.Port,
		}); err != nil {
			return err
		}
		if err := s.applier.ApplyManifest(ctx, kubernetes.NewManifestReader(envoyFilterSpec.Bytes()), kubernetes.DefaultMergeFuncs); err != nil {
			return err
		}
	}

	if _, err := controllerutils.GetAndCreateOrMergePatch(ctx, s.client, destinationRule, func() error {
		destinationRule.Labels = getLabels()
		destinationRule.Spec = istioapinetworkingv1beta1.DestinationRule{
			ExportTo: []string{"*"},
			Host:     hostName,
			TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
				ConnectionPool: &istioapinetworkingv1beta1.ConnectionPoolSettings{
					Tcp: &istioapinetworkingv1beta1.ConnectionPoolSettings_TCPSettings{
						MaxConnections: 5000,
						TcpKeepalive: &istioapinetworkingv1beta1.ConnectionPoolSettings_TCPSettings_TcpKeepalive{
							Time:     &durationpb.Duration{Seconds: 7200},
							Interval: &durationpb.Duration{Seconds: 75},
						},
					},
				},
				LoadBalancer: &istioapinetworkingv1beta1.LoadBalancerSettings{
					LocalityLbSetting: &istioapinetworkingv1beta1.LocalityLoadBalancerSetting{
						Enabled:          &wrapperspb.BoolValue{Value: true},
						FailoverPriority: []string{corev1.LabelTopologyZone},
					},
				},
				// OutlierDetection is required for locality settings to take effect
				OutlierDetection: &istioapinetworkingv1beta1.OutlierDetection{
					MinHealthPercent: 0,
				},
				Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
					Mode: istioapinetworkingv1beta1.ClientTLSSettings_DISABLE,
				},
			},
		}
		return nil
	}); err != nil {
		return err
	}

	if _, err := controllerutils.GetAndCreateOrMergePatch(ctx, s.client, gateway, func() error {
		gateway.Labels = getLabels()
		gateway.Spec = istioapinetworkingv1beta1.Gateway{
			Selector: s.valuesFunc().IstioIngressGateway.Labels,
			Servers: []*istioapinetworkingv1beta1.Server{{
				Hosts: s.valuesFunc().Hosts,
				Port: &istioapinetworkingv1beta1.Port{
					Number:   kubeapiserverconstants.Port,
					Name:     "tls",
					Protocol: "TLS",
				},
				Tls: &istioapinetworkingv1beta1.ServerTLSSettings{
					Mode: istioapinetworkingv1beta1.ServerTLSSettings_PASSTHROUGH,
				},
			}},
		}
		return nil
	}); err != nil {
		return err
	}

	if _, err := controllerutils.GetAndCreateOrMergePatch(ctx, s.client, virtualService, func() error {
		virtualService.Labels = getLabels()
		virtualService.Spec = istioapinetworkingv1beta1.VirtualService{
			ExportTo: []string{"*"},
			Hosts:    s.valuesFunc().Hosts,
			Gateways: []string{gateway.Name},
			Tls: []*istioapinetworkingv1beta1.TLSRoute{{
				Match: []*istioapinetworkingv1beta1.TLSMatchAttributes{{
					Port:     kubeapiserverconstants.Port,
					SniHosts: s.valuesFunc().Hosts,
				}},
				Route: []*istioapinetworkingv1beta1.RouteDestination{{
					Destination: &istioapinetworkingv1beta1.Destination{
						Host: hostName,
						Port: &istioapinetworkingv1beta1.PortSelector{Number: kubeapiserverconstants.Port},
					},
				}},
			}},
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *sni) Destroy(ctx context.Context) error {
	return kubernetesutils.DeleteObjects(
		ctx,
		s.client,
		s.emptyDestinationRule(),
		s.emptyEnvoyFilter(),
		s.emptyGateway(),
		s.emptyVirtualService(),
	)
}

func (s *sni) Wait(_ context.Context) error        { return nil }
func (s *sni) WaitCleanup(_ context.Context) error { return nil }

func (s *sni) emptyDestinationRule() *istionetworkingv1beta1.DestinationRule {
	return &istionetworkingv1beta1.DestinationRule{ObjectMeta: metav1.ObjectMeta{Name: s.name, Namespace: s.namespace}}
}

func (s *sni) emptyEnvoyFilter() *istionetworkingv1alpha3.EnvoyFilter {
	return &istionetworkingv1alpha3.EnvoyFilter{ObjectMeta: metav1.ObjectMeta{Name: s.namespace, Namespace: s.valuesFunc().IstioIngressGateway.Namespace}}
}

func (s *sni) emptyGateway() *istionetworkingv1beta1.Gateway {
	return &istionetworkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: s.name, Namespace: s.namespace}}
}

func (s *sni) emptyVirtualService() *istionetworkingv1beta1.VirtualService {
	return &istionetworkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: s.name, Namespace: s.namespace}}
}

// AnyDeployedSNI returns true if any SNI is deployed in the cluster.
func AnyDeployedSNI(ctx context.Context, c client.Client) (bool, error) {
	l := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": istionetworkingv1beta1.SchemeGroupVersion.String(),
			"kind":       "VirtualServiceList",
		},
	}

	if err := c.List(ctx, l, client.MatchingFields{"metadata.name": "kube-apiserver"}, client.Limit(1)); err != nil && !meta.IsNoMatchError(err) {
		return false, err
	}

	return len(l.Items) > 0, nil
}
