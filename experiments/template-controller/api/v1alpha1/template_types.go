/*


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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TemplateSpec defines the desired state of Template
type TemplateSpec struct {
	// Expander specifies the underlying template engine to use
	// +kubebuilder:validation:Enum=basic
	Expander string `json:"expander,omitempty"`

	ReconcilerFor ReconcilerFor `json:"actsOn"`

	// Resources is a list of Resource
	Resources []Resource `json:"resources,omitempty"`
}

type ReconcilerFor struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

type Resource struct {
	// Name of the resource
	Name string `json:"name"`

	// Template contains inline YAML
	Template runtime.RawExtension `json:"template"`

	// Conditions is a string that describes conditions
	Conditions string `json:"conditions,omitempty"`

	// DependsOn is a list of names of resources that this resource depends on
	DependsOn []string `json:"dependsOn,omitempty"`
}

// TemplateStatus defines the observed state of Template
type TemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Template is the Schema for the templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
