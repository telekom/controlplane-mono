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

// ConsumeRouteSpec defines the desired state of ConsumeRoute
type ConsumeRouteSpec struct {
	Route        types.ObjectRef `json:"route"`
	ConsumerName string          `json:"consumerName"`
}

// ConsumeRouteStatus defines the observed state of ConsumeRoute
type ConsumeRouteStatus struct {
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ConsumeRoute is the Schema for the consumeroutes API
type ConsumeRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsumeRouteSpec   `json:"spec,omitempty"`
	Status ConsumeRouteStatus `json:"status,omitempty"`
}

var _ types.Object = &ConsumeRoute{}

func (c *ConsumeRoute) GetConditions() []metav1.Condition {
	return c.Status.Conditions
}

func (c *ConsumeRoute) SetCondition(condition metav1.Condition) bool {
	return meta.SetStatusCondition(&c.Status.Conditions, condition)
}

// +kubebuilder:object:root=true

// ConsumeRouteList contains a list of ConsumeRoute
type ConsumeRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConsumeRoute `json:"items"`
}

var _ types.ObjectList = &ConsumeRouteList{}

func (c *ConsumeRouteList) GetItems() []types.Object {
	items := make([]types.Object, len(c.Items))
	for i := range c.Items {
		items[i] = &c.Items[i]
	}
	return items
}

func init() {
	SchemeBuilder.Register(&ConsumeRoute{}, &ConsumeRouteList{})
}
