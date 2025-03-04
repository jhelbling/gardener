// Copyright 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package clusteropenidconnectpreset

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"

	"github.com/gardener/gardener/pkg/api"
	"github.com/gardener/gardener/pkg/apis/settings"
	"github.com/gardener/gardener/pkg/apis/settings/validation"
)

type clusterOIDCPresetStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy defines the storage strategy for clusteropenidconnectpresets.
var Strategy = clusterOIDCPresetStrategy{api.Scheme, names.SimpleNameGenerator}

func (clusterOIDCPresetStrategy) NamespaceScoped() bool {
	return false
}

func (clusterOIDCPresetStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {

}

func (clusterOIDCPresetStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {

}

func (clusterOIDCPresetStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	oidcpreset := obj.(*settings.ClusterOpenIDConnectPreset)
	return validation.ValidateClusterOpenIDConnectPreset(oidcpreset)
}

func (clusterOIDCPresetStrategy) Canonicalize(obj runtime.Object) {
}

func (clusterOIDCPresetStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (clusterOIDCPresetStrategy) ValidateUpdate(ctx context.Context, newObj, oldObj runtime.Object) field.ErrorList {
	newOIDCPreset := newObj.(*settings.ClusterOpenIDConnectPreset)
	oldOIDCPreset := oldObj.(*settings.ClusterOpenIDConnectPreset)
	return validation.ValidateClusterOpenIDConnectPresetUpdate(newOIDCPreset, oldOIDCPreset)
}

func (clusterOIDCPresetStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// WarningsOnCreate returns warnings to the client performing a create.
func (clusterOIDCPresetStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// WarningsOnUpdate returns warnings to the client performing the update.
func (clusterOIDCPresetStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
