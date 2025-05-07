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
	"net/url"
	"path"

	"github.com/telekom/controlplane-mono/common/pkg/types"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RealmSpec defines the desired state of Realm
type RealmSpec struct {

	// Gateway is the Gateway that is associated with the Realm
	// If empty, the realm considered a virtual realm
	Gateway *types.ObjectRef `json:"gateway,omitempty"`
	// Url is the URL of the Gateway when its used as an upstream
	Url string `json:"url"`
	// IssuerUrl is the URL of the issuer of the Token sent by the Gateway when its used as a downstream
	IssuerUrl string `json:"issuerUrl"`
	// DefaultConsumers is a list of consumers that are allowed to access the Realm for all Routes
	// +listType=set
	// +kubebuilder:default={}
	DefaultConsumers []string `json:"defaultConsumers"`
}

// RealmStatus defines the observed state of Realm
type RealmStatus struct {
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Virtual indicates that this is a virtual realm and has no Gateway associated with it
	Virtual bool `json:"virtual"`

	IssuerRoute    *types.ObjectRef `json:"issuerRoute,omitempty"`
	IssuerUrl      string           `json:"issuerUrl,omitempty"`
	CertsRoute     *types.ObjectRef `json:"certsRoute,omitempty"`
	CertsUrl       string           `json:"certsUrl,omitempty"`
	DiscoveryRoute *types.ObjectRef `json:"discoveryRoute,omitempty"`
	DiscoveryUrl   string           `json:"discoveryUrl,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Realm is the Schema for the realms API
type Realm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RealmSpec   `json:"spec,omitempty"`
	Status RealmStatus `json:"status,omitempty"`
}

var _ types.Object = &Realm{}

func (r *Realm) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *Realm) SetCondition(condition metav1.Condition) bool {
	return meta.SetStatusCondition(&r.Status.Conditions, condition)
}

func (r *Realm) AsUpstream(apiBasePath string) (ups Upstream, err error) {
	url, err := url.Parse(r.Spec.Url)
	if err != nil {
		return ups, err
	}
	ups = Upstream{
		Scheme: url.Scheme,
		Host:   url.Hostname(),
		Port:   GetPortOrDefaultFromScheme(url),
		Path:   path.Join(url.Path, apiBasePath),
	}
	return
}

func (r *Realm) AsDownstream(apiBasePath string) (dws Downstream, err error) {
	url, err := url.Parse(r.Spec.Url)
	if err != nil {
		return dws, err
	}
	dws = Downstream{
		Host:      url.Hostname(),
		Port:      GetPortOrDefaultFromScheme(url),
		Path:      path.Join(url.Path, apiBasePath),
		IssuerUrl: r.Spec.IssuerUrl,
	}
	return
}

// +kubebuilder:object:root=true

// RealmList contains a list of Realm
type RealmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Realm `json:"items"`
}

var _ types.ObjectList = &RealmList{}

func (l *RealmList) GetItems() []types.Object {
	items := make([]types.Object, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}

func init() {
	SchemeBuilder.Register(&Realm{}, &RealmList{})
}
