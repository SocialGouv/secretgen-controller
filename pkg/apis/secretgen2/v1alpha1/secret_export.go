// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"

	sgv1alpha1 "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name=Description,JSONPath=.status.friendlyDescription,description=Friendly description,type=string
// +kubebuilder:printcolumn:name=Age,JSONPath=.metadata.creationTimestamp,description=Time since creation,type=date
type SecretExport struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata; More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SecretExportSpec `json:"spec"`
	// +optional
	Status SecretExportStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SecretExportList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SecretExport `json:"items"`
}

type SecretExportSpec struct {
	// +optional
	ToNamespace string `json:"toNamespace,omitempty"`
	// +optional
	ToNamespaces []string `json:"toNamespaces,omitempty"`
	// +optional
	ToNamespaceAnnotation map[string]string `json:"toNamespaceAnnotation,omitempty"`
	// +optional
	ToNamespaceAnnotations map[string][]string `json:"toNamespaceAnnotations,omitempty"`
}

type SecretExportStatus struct {
	sgv1alpha1.GenericStatus `json:",inline"`
	// +optional
	ObservedSecretResourceVersion string `json:"observedSecretResourceVersion,omitempty"`
}

// Annotation key and value pair matching a namespace
type SecretExportAnnotation struct {
	Key   string
	Value string
}

const (
	AllNamespaces = "*"
)

func (e SecretExport) StaticToNamespaces() []string {
	result := append([]string{}, e.Spec.ToNamespaces...)
	if len(e.Spec.ToNamespace) > 0 {
		result = append(result, e.Spec.ToNamespace)
	}
	return result
}

// Aggregate ToNamespaceAnnotation and ToNamespaceAnnotations as a single slice
func (e SecretExport) StaticToNamespacesAnnotations() []*SecretExportAnnotation {
	var result []*SecretExportAnnotation
	for k, v := range e.Spec.ToNamespaceAnnotation {
		result = append(result, &SecretExportAnnotation{
			Key:   k,
			Value: v,
		})
	}
	for k, v := range e.Spec.ToNamespaceAnnotations {
		for _, value := range v {
			result = append(result, &SecretExportAnnotation{
				Key:   k,
				Value: value,
			})
		}
	}
	return result
}

func (e SecretExport) Validate() error {
	var errs []error

	toNses := e.StaticToNamespaces()
	toNsesA := e.StaticToNamespacesAnnotations()

	if len(toNses) == 0 && len(toNsesA) == 0 {
		errs = append(errs, fmt.Errorf("Expected to have at least one non-empty to namespace or to namespace annotation"))
	}
	for _, ns := range toNses {
		if len(ns) == 0 {
			errs = append(errs, fmt.Errorf("Expected to namespace to be non-empty"))
		}
	}

	return combinedErrs("Validation errors", errs)
}
