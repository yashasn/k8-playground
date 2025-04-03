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

package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-redis/redis/v8"
	databasev1alpha1 "github.com/yashasn/database-operator/api/v1alpha1"
)

// DatabaseBackupReconciler reconciles a DatabaseBackup object
type DatabaseBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.example.com,resources=databasebackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.example.com,resources=databasebackups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.example.com,resources=databasebackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DatabaseBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation", "request", req)

	var backup databasev1alpha1.DatabaseBackup
	if err := r.Get(ctx, req.NamespacedName, &backup); err != nil {
		logger.Error(err, "Unable to fetch DatabaseBackup resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Access the Redis credentials from the Kubernetes secret
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: backup.Spec.ConnectionSecret}, secret)
	if err != nil {
		logger.Error(err, "Unable to fetch the Redis credentials secret")
		return ctrl.Result{}, err
	}

	redisPasswordBase64 := string(secret.Data["redis-password"])
	logger.Info("base64 password", "password", string(redisPasswordBase64))
	// redisPassword, err := base64.StdEncoding.DecodeString(redisPasswordBase64)
	// if err != nil {
	// 	logger.Error(err, "Failed to decode Redis password")
	// 	return ctrl.Result{}, err
	// }

	// Get Redis address and backup directory from spec
	redisAddr := backup.Spec.RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default value
	}

	backupDir := backup.Spec.BackupDirectory
	if backupDir == "" {
		backupDir = "/tmp/db_backups" // Default value
	}

	backupInterval := backup.Spec.BackupInterval
	if backupInterval == "" {
		backupInterval = "5m" // Default to 5 minutes
	}

	// Convert the backup interval to time.Duration
	duration, err := time.ParseDuration(backupInterval)
	if err != nil {
		logger.Error(err, "Invalid backup interval format")
		return ctrl.Result{}, fmt.Errorf("invalid backup interval: %v", err)
	}

	// Create Redis client
	logger.Info("Connecting to Redis", "redisAddr", redisAddr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPasswordBase64, // Password from secret
		DB:       0,                   // Default DB
	})

	// Fetch all keys from Redis
	logger.Info("Fetching keys from Redis")
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		logger.Error(err, "Failed to fetch keys from Redis")
		return ctrl.Result{}, err
	}

	// Create backup directory if not exists
	logger.Info("Ensuring backup directory exists", "backupDir", backupDir)
	err = os.MkdirAll(backupDir, os.ModePerm)
	if err != nil {
		logger.Error(err, "Failed to create backup directory")
		return ctrl.Result{}, err
	}

	// Back up the data to a text file
	var data string
	for _, key := range keys {
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			logger.Error(err, "Failed to retrieve value from Redis", "key", key)
			continue
		}
		data += fmt.Sprintf("%s: %s\n", key, val)
	}

	// Write data to a backup file
	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%d.txt", time.Now().Unix()))
	err = os.WriteFile(backupFile, []byte(data), 0644)
	if err != nil {
		logger.Error(err, "Failed to write backup file", "backupFile", backupFile)
		return ctrl.Result{}, err
	}

	logger.Info("Backup successfully completed", "backupFile", backupFile)

	// Update CRD status
	backup.Status.Phase = "Completed"
	backup.Status.LastBackupTime = metav1.Now()

	if err := r.Status().Update(ctx, &backup); err != nil {
		logger.Error(err, "Failed to update DatabaseBackup status")
		return ctrl.Result{}, err
	}

	// Requeue based on the backupInterval (configured duration)
	logger.Info("Reconciliation completed, requeuing", "requeueAfter", duration)
	return ctrl.Result{RequeueAfter: duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1alpha1.DatabaseBackup{}).
		Named("databasebackup").
		Complete(r)
}
