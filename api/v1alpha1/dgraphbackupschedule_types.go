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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DgraphBackupScheduleSpec defines the desired state of DgraphBackupSchedule
type DgraphBackupScheduleSpec struct {
	// Schedule is schedule info in github.com/robfig/cron supported notation
	Schedule string `json:"schedule"`

	// Retention is specify how long should to keep backups
	Retention string `json:"retention,omitempty"`

	// Backup is specify dgraph backup options
	Backup DgraphBackupSpec `json:"backup"`
}

// DgraphBackupScheduleStatus defines the observed state of DgraphBackupSchedule
type DgraphBackupScheduleStatus struct {
	ScheduleTaskID   int         `json:"scheduleTaskId,omitempty"`
	RetentionTaskID  int         `json:"retentionTaskId,omitempty"`
	ActiveGeneration int64       `json:"activeGeneration,omitempty"`
	UpdatedAt        metav1.Time `json:"updatedTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule",description="backup objects creation schedule"
//+kubebuilder:printcolumn:name="Retention",type="string",JSONPath=".spec.retention",description="backup objects retention perion"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// DgraphBackupSchedule is the Schema for the dgraphbackupschedules API
type DgraphBackupSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DgraphBackupScheduleSpec   `json:"spec,omitempty"`
	Status DgraphBackupScheduleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DgraphBackupScheduleList contains a list of DgraphBackupSchedule
type DgraphBackupScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DgraphBackupSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DgraphBackupSchedule{}, &DgraphBackupScheduleList{})
}

func (cr *DgraphBackupSchedule) AsOwner() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         cr.APIVersion,
			Kind:               cr.Kind,
			Name:               cr.Name,
			UID:                cr.UID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		},
	}
}

func (cr DgraphBackupSchedule) Annotations() map[string]string {
	annotations := make(map[string]string)
	for annotation, value := range cr.ObjectMeta.Annotations {
		if !strings.HasPrefix(annotation, "kubectl.kubernetes.io/") {
			annotations[annotation] = value
		}
	}
	return annotations
}

// IsNeedUpdate returns true if resource must be updated
func (s *DgraphBackupSchedule) IsNeedUpdate(startedAt *metav1.Time) bool {
	if s.Generation != s.Status.ActiveGeneration {
		return true
	}

	if s.Status.UpdatedAt.Before(startedAt) {
		return true
	}

	return false
}
