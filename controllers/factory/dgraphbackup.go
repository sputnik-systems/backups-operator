package factory

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
	"github.com/sputnik-systems/backups-operator/controllers/factory/finalize"
	"github.com/sputnik-systems/backups-operator/internal/dgraph"
)

func ProccessDgraphBackupObject(ctx context.Context, rc client.Client, b *backupsv1alpha1.DgraphBackup) error {
	if b.Status.Phase == "" {
		if err := finalize.AddFinalizer(ctx, rc, b); err != nil {
			return fmt.Errorf("failed to add finalizer: %w", err)
		}

		b.Status.Phase = "Started"
		if err := rc.Status().Update(ctx, b); err != nil {
			return fmt.Errorf("failed update status: %w", err)
		}

		if err := createDgraphBackup(ctx, rc, b); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		b.Status.Phase = "Completed"
		if err := rc.Status().Update(ctx, b); err != nil {
			return fmt.Errorf("failed update status: %w", err)
		}
	}

	return nil
}

func DeleteDgraphBackupObject(ctx context.Context, rc client.Client, b *backupsv1alpha1.DgraphBackup) error {
	creds, err := getCredentials(ctx, rc, b.Spec.Secrets, b.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get creds: %w", err)
	}

	if err := dgraph.DeleteExport(ctx, b, creds); err != nil {
		return fmt.Errorf("failed to delete backup from remote storage: %w", err)
	}

	if err := finalize.RemoveFinalizeObjByName(ctx, rc, b, b.Name, b.Namespace); err != nil {
		return fmt.Errorf("failed to remove finalizer: %w", err)
	}

	return nil
}

func createDgraphBackup(ctx context.Context, rc client.Client, b *backupsv1alpha1.DgraphBackup) error {
	creds, err := getCredentials(ctx, rc, b.Spec.Secrets, b.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get dgraph export creds: %w", err)
	}

	b.Spec.AdminUrl, err = getFQDN(b.Spec.AdminUrl, b.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get fqdn: %w", err)
	}

	out, err := dgraph.Export(ctx, rc, &b.Spec, creds)
	if err != nil {
		b.Status.Phase = "Failed"
		if err := rc.Status().Update(ctx, b); err != nil {
			return fmt.Errorf("failed to update dgraph backup object status: %w", err)
		}

		return err
	}

	files := make([]string, 0)
	for _, file := range out.ExportedFiles {
		files = append(files, string(file))
	}

	b.Status.ExportResponse.ExportedFiles = files
	b.Status.ExportResponse.Message = string(out.Response.Message)
	b.Status.ExportResponse.Code = string(out.Response.Code)

	return rc.Status().Update(ctx, b)
}
