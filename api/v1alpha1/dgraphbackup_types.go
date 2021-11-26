/*
Copyright 2021.

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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DgraphBackupSpec defines the desired state of DgraphBackup
type DgraphBackupSpec struct {
	// AdminUrl is dgraph alpha instance admin url
	AdminUrl string `json:"adminUrl"`

	// Namespace is dgraph exported namespace
	Namespace int `json:"namespace,omitempty"`

	// Format is dgraph export file format
	Format string `json:"format,omitempty"`

	// Dest is backup destination
	Destination string `json:"destination"`

	// Region is s3 storage region
	Region string `json:"region,omitempty"`

	// Secrets is list of secret abstraction names
	Secrets []string `json:"secrets,omitempty"`

	// Anonymous if credentials is not required
	Anonymous bool `json:"anonymous,omitempty"`
}

// DgraphBackupStatus defines the observed state of DgraphBackup
type DgraphBackupStatus struct {
	Phase          string                          `json:"phase,omitempty"`
	ExportResponse DgraphBackupStatusExportResonse `json:"exportResponse,omitempty"`
}

type DgraphBackupStatusExportResonse struct {
	Message       string   `json:"message,omitempty"`
	Code          string   `json:"code,omitempty"`
	ExportedFiles []string `json:"exportedFiles,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="backup creation phase"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// DgraphBackup is the Schema for the dgraphbackups API
type DgraphBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DgraphBackupSpec   `json:"spec,omitempty"`
	Status DgraphBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DgraphBackupList contains a list of DgraphBackup
type DgraphBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DgraphBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DgraphBackup{}, &DgraphBackupList{})
}
