/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"github.com/telekom/controlplane-mono/common/pkg/types"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IdentityProviderSpec defines the desired state of IdentityProvider
type IdentityProviderSpec struct {
	AdminUrl      string `json:"adminUrl"`
	AdminClientId string `json:"adminClientId"`
	AdminUserName string `json:"adminUserName"`
	AdminPassword string `json:"adminPassword"`
}

// IdentityProviderStatus defines the observed state of IdentityProvider
type IdentityProviderStatus struct {
	// Expected format for the admin URL is https://<host>/auth/admin/realms/
	// Example: https://iris-distcp1-dataplane1.dev.dhei.telekom.de/auth/admin/realms/
	AdminUrl        string `json:"adminUrl"`
	AdminTokenUrl   string `json:"adminTokenUrl"`
	AdminConsoleUrl string `json:"adminConsoleUrl,omitempty"`
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IdentityProvider is the Schema for the identityproviders API
type IdentityProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityProviderSpec   `json:"spec,omitempty"`
	Status IdentityProviderStatus `json:"status,omitempty"`
}

var _ types.Object = &IdentityProvider{}

func (e *IdentityProvider) GetConditions() []metav1.Condition {
	return e.Status.Conditions
}

func (e *IdentityProvider) SetCondition(condition metav1.Condition) bool {
	return meta.SetStatusCondition(&e.Status.Conditions, condition)
}

// +kubebuilder:object:root=true

// IdentityProviderList contains a list of IdentityProvider
type IdentityProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityProvider `json:"items"`
}

var _ types.ObjectList = &IdentityProviderList{}

func (el *IdentityProviderList) GetItems() []types.Object {
	items := make([]types.Object, len(el.Items))
	for i := range el.Items {
		items[i] = &el.Items[i]
	}
	return items
}

func init() {
	SchemeBuilder.Register(&IdentityProvider{}, &IdentityProviderList{})
}
