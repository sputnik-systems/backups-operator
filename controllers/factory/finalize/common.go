package finalize

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
)

// AddFinalizer adds finalizer to obj if needed.
func AddFinalizer(ctx context.Context, rc client.Client, obj client.Object) error {
	if !backupsv1alpha1.IsContainsFinalizer(obj.GetFinalizers(), backupsv1alpha1.FinalizerName) {
		obj.SetFinalizers(append(obj.GetFinalizers(), backupsv1alpha1.FinalizerName))
		return rc.Update(ctx, obj)
	}
	return nil
}

// RemoveFinalizer removes finalizer from obj if needed.
func RemoveFinalizer(ctx context.Context, rc client.Client, obj client.Object) error {
	if backupsv1alpha1.IsContainsFinalizer(obj.GetFinalizers(), backupsv1alpha1.FinalizerName) {
		obj.SetFinalizers(backupsv1alpha1.RemoveFinalizer(obj.GetFinalizers(), backupsv1alpha1.FinalizerName))
		return rc.Update(ctx, obj)
	}
	return nil
}

func RemoveFinalizeObjByName(ctx context.Context, rc client.Client, obj client.Object, name, ns string) error {
	if err := rc.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, obj); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	obj.SetFinalizers(backupsv1alpha1.RemoveFinalizer(obj.GetFinalizers(), backupsv1alpha1.FinalizerName))

	return rc.Update(ctx, obj)
}
