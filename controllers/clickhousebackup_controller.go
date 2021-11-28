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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
	"github.com/sputnik-systems/backups-operator/controllers/factory"
	"github.com/sputnik-systems/backups-operator/controllers/factory/finalize"
	"github.com/sputnik-systems/backups-operator/internal/clickhouse"
)

// ClickHouseBackupReconciler reconciles a ClickHouseBackup object
type ClickHouseBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=clickhousebackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=clickhousebackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=clickhousebackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClickHouseBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ClickHouseBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("started resource reconclie")

	b := &backupsv1alpha1.ClickHouseBackup{}
	err := r.Get(ctx, req.NamespacedName, b)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		l.Error(err, "failed to get clickhouse backup object for reconclie")
		return ctrl.Result{}, err
	}

	if !b.DeletionTimestamp.IsZero() {
		if _, err := clickhouse.DeleteBackup(ctx, b); err != nil {
			l.Error(err, "failed to delete backup")
		}

		if err := finalize.RemoveFinalizeObjByName(ctx, r.Client, b, b.Name, b.Namespace); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if b.Status.Phase == "" {
		b.Status.Phase = "Started"
		if err := r.Status().Update(ctx, b); err != nil {
			l.Error(err, "failed update clickhouse backup object")

			return ctrl.Result{}, err
		}

		if err := finalize.AddFinalizer(ctx, r.Client, b); err != nil {
			l.Error(err, "failed to add finalizer")

			return ctrl.Result{}, err
		}

		if err := factory.UpdateClickHouseBackupStatusApiInfo(ctx, r.Client, b); err != nil {
			l.Error(err, "failed to update status api info")

			return ctrl.Result{}, err
		}

		if err := factory.CreateClickHouseBackup(ctx, r.Client, b); err != nil {
			l.Error(err, "failed to create backup")

			return ctrl.Result{}, err
		}

		if b.Status.Phase == "Created" {
			if err := factory.UploadClickHouseBackup(ctx, r.Client, b); err != nil {
				l.Error(err, "failed to upload backup")

				return ctrl.Result{}, err
			}
		}

		l.Info("backup created succesfully")
	}

	l.Info("finished resource reconclie")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClickHouseBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupsv1alpha1.ClickHouseBackup{}).
		Complete(r)
}
