package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OceanToolkitSpec defines the desired state of Ocean Toolkit.
type OceanToolkitSpec struct {
	// List of components to install.
	Components map[string]*OceanToolkitComponent `json:"components,omitempty"`
	// Global values. This is a validated pass-through to Helm templates.
	// See the Helm charts for schema details: https://github.com/spotinst/ocsean-charts/.
	Values map[string]interface{} `json:"values,omitempty"`
}

// OceanToolkitComponent defines the desired state of a component.
type OceanToolkitComponent struct {
	// Selects whether this component is installed.
	Enabled bool `json:"enabled"`
	// Local values. This is a validated pass-through to Helm templates.
	// See the Helm charts for schema details: https://github.com/spotinst/ocsean-charts/.
	Values map[string]interface{} `json:"values,omitempty"`
}

// InstallStatus describes the current installation state of a component.
type InstallStatus int32

const (
	// Component is not present.
	InstallStatusUnknown InstallStatus = iota
	// Component is being updated to a different version.
	InstallStatusUpdating
	// Component is being reconciled.
	InstallStatusReconciling
	// Component is healthy.
	InstallStatusHealthy
	// Component is in an error state.
	InstallStatusError
)

// OceanToolkitStatus defines the observed status of Ocean Toolkit.
type OceanToolkitStatus struct {
	// Most recently observed status of the Toolkit.
	ReconcileStatus `json:",inline"`
	// Observed status of the installation.
	InstallStatus map[string]InstallStatus `json:"installStatus,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OceanToolkit consists of one or more Toolkit components.
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusters,scope=Namespaced
type OceanToolkit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the OceanToolkit.
	Spec OceanToolkitSpec `json:"spec,omitempty"`
	// Most recently observed status of the OceanToolkit.
	Status OceanToolkitStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OceanToolkitList contains a list of OceanToolkit.
type OceanToolkitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of OceanToolkits.
	Items []OceanToolkit `json:"items"`
}
