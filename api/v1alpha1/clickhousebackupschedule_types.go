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

// ClickHouseBackupScheduleSpec defines the desired state of ClickHouseBackupSchedule
type ClickHouseBackupScheduleSpec struct {
	// Schedule is schedule info in github.com/robfig/cron supported notation
	Schedule string `json:"schedule"`

	// Retention is specify how long should to keep backups
	Retention string `json:"retention,omitempty"`

	// Backup is specify clickhouse backup options
	Backup ClickHouseBackupSpec `json:"backup"`
}

// ClickHouseBackupScheduleStatus defines the observed state of ClickHouseBackupSchedule
type ClickHouseBackupScheduleStatus struct {
	ScheduleTaskID   int         `json:"scheduleTaskId,omitempty"`
	RetentionTaskID  int         `json:"retentionTaskId,omitempty"`
	ActiveGeneration int64       `json:"activeGeneration,omitempty"`
	UpdatedAt        metav1.Time `json:"updatedTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule",description="backup objects creation schedule"
//+kubebuilder:printcolumn:name="Retention",type="string",JSONPath=".spec.retention",description="backup objects retention period"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ClickHouseBackupSchedule is the Schema for the clickhousebackupschedules API
type ClickHouseBackupSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClickHouseBackupScheduleSpec   `json:"spec,omitempty"`
	Status ClickHouseBackupScheduleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClickHouseBackupScheduleList contains a list of ClickHouseBackupSchedule
type ClickHouseBackupScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClickHouseBackupSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClickHouseBackupSchedule{}, &ClickHouseBackupScheduleList{})
}

func (cr *ClickHouseBackupSchedule) AsOwner() []metav1.OwnerReference {
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

func (cr ClickHouseBackupSchedule) Annotations() map[string]string {
	annotations := make(map[string]string)
	for annotation, value := range cr.ObjectMeta.Annotations {
		if !strings.HasPrefix(annotation, "kubectl.kubernetes.io/") {
			annotations[annotation] = value
		}
	}
	return annotations
}

// IsNeedUpdate returns true if resource must be updated
func (cr *ClickHouseBackupSchedule) IsNeedUpdate(startedAt *metav1.Time) bool {
	if cr.Generation != cr.Status.ActiveGeneration {
		return true
	}

	if cr.Status.UpdatedAt.Before(startedAt) {
		return true
	}

	return false
}
