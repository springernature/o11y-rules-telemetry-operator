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

package v1beta1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type GroupRulesSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	GroupName string               `json:"name"`
	Rules     apiextensionsv1.JSON `json:"rules"`
}

// MimirRulesSpec defines the desired state of MimirRules
type MimirRulesSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	Groups []GroupRulesSpec `json:"groups"`
}

// MimirRulesStatus defines the observed state of MimirRules
type MimirRulesStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	GroupStatus map[string]string `json:"GroupsStatus"`
	Errors      int               `json:"Errors"`
	Tenant      string            `json:"Tenant"`
	LastUpdate  string            `json:"LastUpdate"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MimirRules is the Schema for the mimirrules API
type MimirRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MimirRulesSpec   `json:"spec,omitempty"`
	Status MimirRulesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MimirRulesList contains a list of MimirRules
type MimirRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MimirRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MimirRules{}, &MimirRulesList{})
}
