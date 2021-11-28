package factory

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
	"github.com/sputnik-systems/backups-operator/internal/dgraph"
)

func CreateDgraphBackup(ctx context.Context, rc client.Client, b *backupsv1alpha1.DgraphBackup) error {
	creds, err := GetCredentials(ctx, rc, b.Spec.Secrets, b.Namespace)
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
