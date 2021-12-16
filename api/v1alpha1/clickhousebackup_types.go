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

// ClickHouseBackupSpec defines the desired state of ClickHouseBackup
type ClickHouseBackupSpec struct {
	// ApiAddress is requests sending endpoint
	ApiAddress string `json:"apiAddress"`

	// ExponentialBackOff is specify exponential backoff time settings for backup creation flow
	ExponentialBackOff *ExponentialBackOffSpec `json:"exponentialBackOff,omitempty"`

	// CreateParams is optional backup creating query params
	CreateParams map[string]string `json:"createParams,omitempty"`

	// UploadParams is optional backup uploading query params
	UploadParams map[string]string `json:"uploadParams,omitempty"`
}

// ClickHouseBackupStatus defines the observed state of ClickHouseBackup
type ClickHouseBackupStatus struct {
	// Phase is current state of underlying operation
	Phase string `json:"phase,omitempty"`

	// Api is specify where requests will be send
	Api ClickHouseBackupStatusApi `json:"api,omitempty"`

	// Error is error message if backup creationg failed
	Error string `json:"error,omitempty"`
}

type ClickHouseBackupStatusApi struct {
	// Address is real address for sending requests
	Address string `json:"Address,omitempty"`

	// Hostname is Hostname header value
	Hostname string `json:"Hostname,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="backup creation phase"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ClickHouseBackup is the Schema for the clickhousebackups API
type ClickHouseBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClickHouseBackupSpec   `json:"spec,omitempty"`
	Status ClickHouseBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClickHouseBackupList contains a list of ClickHouseBackup
type ClickHouseBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClickHouseBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClickHouseBackup{}, &ClickHouseBackupList{})
}
