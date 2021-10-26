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
	"github.com/sputnik-systems/backups-operator/controllers/factory/finalize"
)

// DgraphBackupScheduleReconciler reconciles a DgraphBackupSchedule object
type DgraphBackupScheduleReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Cron      *cron.Cron
	StartedAt metav1.Time
}

//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=dgraphbackupschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=dgraphbackupschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=backups.sputnik.systems,resources=dgraphbackupschedules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DgraphBackupSchedule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *DgraphBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("started resource reconclie")

	bs := &backupsv1alpha1.DgraphBackupSchedule{}
	err := r.Get(ctx, req.NamespacedName, bs)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		l.Error(err, "failed to get dgraph backupi schedule object for reconclie")
		return ctrl.Result{}, err
	}

	if !bs.DeletionTimestamp.IsZero() {
		id := cron.EntryID(bs.Status.ScheduleID)
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

	if r.isNeedUpdate(ctx, bs) {
		// if err := r.updateStatusState(ctx, bs, "Started"); err != nil {
		// 	l.Error(err, "failed update dgraph backup schedule object")

		// 	return ctrl.Result{}, err
		// }

		if err := finalize.AddFinalizer(ctx, r.Client, bs); err != nil {
			l.Error(err, "failed to add finalizer")

			return ctrl.Result{}, err
		}

		var id cron.EntryID
		if bs.Status.ScheduleID != 0 {
			id = cron.EntryID(bs.Status.ScheduleID)

			for _, entry := range r.Cron.Entries() {
				if entry.ID == id {
					r.Cron.Remove(id)
				}
			}
		}

		id, err = r.Cron.AddFunc(bs.Spec.Schedule, func() { r.schedule(ctx, bs) })
		if err != nil {
			l.Error(err, "failed to schedule dgraph backup")

			return ctrl.Result{}, err
		}

		bs.Status.ScheduleID = int(id)
		// bs.Status.State = "Completed"
		bs.Status.ActiveGeneration = bs.Generation
		bs.Status.UpdatedAt = metav1.Now()

		if err := r.Status().Update(ctx, bs); err != nil {
			l.Error(err, "failed update dgraph backup schedule object")

			return ctrl.Result{}, err
		}
	}

	l.Info("finished resource reconclie")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DgraphBackupScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupsv1alpha1.DgraphBackupSchedule{}).
		Complete(r)
}

// func (r *DgraphBackupScheduleReconciler) updateStatusState(ctx context.Context, bs *backupsv1alpha1.DgraphBackupSchedule, state string) error {
// 	bs.Status.State = state
//
// 	return r.Status().Update(ctx, bs)
// }

// func (r *DgraphBackupScheduleReconciler) updateStatusID(ctx context.Context, bs *backupsv1alpha1.DgraphBackupSchedule, id int) error {
// 	bs.Status.ScheduleID = id
//
// 	return r.Status().Update(ctx, bs)
// }

func (r *DgraphBackupScheduleReconciler) isNeedUpdate(ctx context.Context, bs *backupsv1alpha1.DgraphBackupSchedule) bool {
	if bs.Generation != bs.Status.ActiveGeneration {
		return true
	}

	if bs.Status.UpdatedAt.Before(&r.StartedAt) {
		return true
	}

	return false
}

func (r *DgraphBackupScheduleReconciler) schedule(ctx context.Context, bs *backupsv1alpha1.DgraphBackupSchedule) {
	l := log.FromContext(ctx)
	name := fmt.Sprintf("%s-%d", bs.Name, time.Now().Unix())

	b := &backupsv1alpha1.DgraphBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: bs.Namespace,
		},
		Spec: bs.Spec.Backup,
	}

	if err := r.Create(ctx, b); err != nil {
		l.Error(err, "failed to create dgraph backup object")
	}
}
