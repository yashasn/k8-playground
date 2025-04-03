/*
Copyright 2025.

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

// DatabaseBackupSpec defines the desired state of DatabaseBackup.
type DatabaseBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DatabaseName     string `json:"databaseName"`
	DatabaseType     string `json:"databaseType"`
	ConnectionSecret string `json:"connectionSecret"`
	StorageType      string `json:"storageType"`
	BackupInterval   string `json:"backupInterval,omitempty"`
	BackupDirectory  string `json:"backupDirectory,omitempty"`
	RedisAddr        string `json:"redisAddr,omitempty"`
}

// DatabaseBackupStatus defines the observed state of DatabaseBackup.
type DatabaseBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase          string      `json:"phase,omitempty"`
	LastBackupTime metav1.Time `json:"lastBackupTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DatabaseBackup is the Schema for the databasebackups API.
type DatabaseBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseBackupSpec   `json:"spec,omitempty"`
	Status DatabaseBackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseBackupList contains a list of DatabaseBackup.
type DatabaseBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseBackup{}, &DatabaseBackupList{})
}
