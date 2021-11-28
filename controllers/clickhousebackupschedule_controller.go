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
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
	"github.com/sputnik-systems/backups-operator/controllers/factory"
	"github.com/sputnik-systems/backups-operator/controllers/factory/finalize"
)

// ClickHouseBackupScheduleReconciler reconciles a ClickHouseBackupSchedule object
type ClickHouseBackupScheduleReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Cron      *cron.Cron
	StartedAt metav1.Time
}

//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=clickhousebackupschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=clickhousebackupschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=clickhousebackupschedules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClickHouseBackupSchedule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ClickHouseBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("started resource reconclie")

	bs := &backupsv1alpha1.ClickHouseBackupSchedule{}
	err := r.Get(ctx, req.NamespacedName, bs)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		l.Error(err, "failed to get clickhouse backupi schedule object for reconclie")
		return ctrl.Result{}, err
	}

	if !bs.DeletionTimestamp.IsZero() {
		id := cron.EntryID(bs.Status.ScheduleTaskID)
		for _, entry := range r.Cron.Entries() {
			if entry.ID == id {
				r.Cron.Remove(id)
			}
		}

		if err := finalize.RemoveFinalizeObjByName(ctx, r.Client, bs, bs.Name, bs.Namespace); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if bs.IsNeedUpdate(&r.StartedAt) {
		if err := finalize.AddFinalizer(ctx, r.Client, bs); err != nil {
			l.Error(err, "failed to add finalizer")

			return ctrl.Result{}, err
		}

		createBackupFunc := func() {
			name := fmt.Sprintf("%s-%d", bs.Name, time.Now().Unix())

			b := &backupsv1alpha1.ClickHouseBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:            name,
					Namespace:       bs.Namespace,
					OwnerReferences: bs.AsOwner(),
				},
				Spec: bs.Spec.Backup,
			}

			if err := r.Create(ctx, b); err != nil {
				l.Error(err, "failed to create clickhouse backup object")
			}
		}

		id, err := factory.ScheduleTask(r.Cron, bs.Spec.Schedule, bs.Status.ScheduleTaskID, createBackupFunc)
		if err != nil {
			l.Error(err, "failed to schedule clickhouse backup")

			return ctrl.Result{}, err
		}

		bs.Status.ScheduleTaskID = int(id)
		bs.Status.ActiveGeneration = bs.Generation
		bs.Status.UpdatedAt = metav1.Now()

		if bs.Spec.Retention != "" {
			rd, err := time.ParseDuration(bs.Spec.Retention)
			if err != nil {
				l.Error(err, "failed to parse retention duration")

				return ctrl.Result{}, err
			}

			removeOutdatedBackupsFunc := func() {
				bl := &backupsv1alpha1.ClickHouseBackupList{}

				if err := r.List(ctx, bl); err != nil {
					l.Error(err, "failed to list clickhouse backup objects")
				}

				for _, item := range bl.Items {
					owner := metav1.GetControllerOf(&item)
					if owner != nil {
						if bs.ObjectMeta.UID != owner.UID {
							continue
						}

						dt := item.ObjectMeta.CreationTimestamp.Time
						if time.Since(dt) > rd {
							l.Info(fmt.Sprintf("delete clickhouse backup %s", item.ObjectMeta.Name))

							if err := r.Delete(ctx, &item); err != nil {
								l.Error(err, "failed to delete clickhouse backup object")
							}
						}
					}
				}
			}

			id, err := factory.ScheduleTask(r.Cron, "@hourly", bs.Status.RetentionTaskID, removeOutdatedBackupsFunc)
			if err != nil {
				l.Error(err, "failed to schedule clickhouse backup")

				return ctrl.Result{}, err
			}

			bs.Status.RetentionTaskID = int(id)
		}

		if err := r.Status().Update(ctx, bs); err != nil {
			l.Error(err, "failed update clickhouse backup schedule object")

			return ctrl.Result{}, err
		}
	}

	l.Info("finished resource reconclie")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClickHouseBackupScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupsv1alpha1.ClickHouseBackupSchedule{}).
		Complete(r)
}
