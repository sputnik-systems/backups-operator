package factory

import (
	"context"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
	"github.com/sputnik-systems/backups-operator/controllers/factory/finalize"
	"github.com/sputnik-systems/backups-operator/internal/clickhouse"
)

func ProccessClickHouseBackupObject(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
	if b.Status.Phase == "" {
		if err := finalize.AddFinalizer(ctx, rc, b); err != nil {
			return fmt.Errorf("failed to add finalizer: %w", err)
		}

		b.Status.Phase = "Started"
		if err := rc.Status().Update(ctx, b); err != nil {
			return fmt.Errorf("failed update status: %w", err)
		}

		if err := updateClickHouseBackupObjectStatusApiInfo(ctx, rc, b); err != nil {
			return fmt.Errorf("failed to update status api info: %w", err)
		}

		if err := createClickHouseBackup(ctx, rc, b); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		if err := uploadClickHouseBackup(ctx, rc, b); err != nil {
			return fmt.Errorf("failed to upload backup: %w", err)
		}
	}

	return nil
}

func DeleteClickHouseBackupObject(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
	if _, err := clickhouse.DeleteBackup(ctx, b); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	if err := finalize.RemoveFinalizeObjByName(ctx, rc, b, b.Name, b.Namespace); err != nil {
		return fmt.Errorf("failed to remove finalizer: %w", err)
	}

	return nil
}

func updateClickHouseBackupObjectStatusApiInfo(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
	var err error

	b.Spec.ApiAddress, err = getFQDN(b.Spec.ApiAddress, b.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get resource fqdn: %w", err)
	}

	b.Status.Api.Address, err = getUrlWithIP(b.Spec.ApiAddress)
	if err != nil {
		return fmt.Errorf("failed to get resource ip address: %w", err)
	}

	b.Status.Api.Hostname, err = getHostname(b.Spec.ApiAddress)
	if err != nil {
		return fmt.Errorf("failed to get resource hostname: %w", err)
	}

	return rc.Status().Update(ctx, b)
}

func createClickHouseBackup(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
	b.Status.Phase = "Creating"
	if err := rc.Status().Update(ctx, b); err != nil {
		return fmt.Errorf("failed update clickhouse backup object: %w", err)
	}

	if _, err := clickhouse.CreateBackup(ctx, b); err != nil {
		b.Status.Phase = "CreateFailed"
		b.Status.Error = err.Error()
		if err := rc.Status().Update(ctx, b); err != nil {
			return err
		}

		return err
	}

	bo := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	op := func() error {
		rows, err := clickhouse.GetStatus(ctx, b)
		if err != nil {
			return fmt.Errorf("failed to get backups status: %w", err)
		}

		if len(rows) > 0 {
			last := rows[len(rows)-1]
			switch last.Status {
			case "error":
				b.Status.Phase = "CreateFailed"
				b.Status.Error = last.Error
				if err := rc.Status().Update(ctx, b); err != nil {
					return backoff.Permanent(err)
				}

				return backoff.Permanent(errors.New("clickhouse backup creating failed"))
			case "success":
				b.Status.Phase = "Created"
				return rc.Status().Update(ctx, b)
			default:
				return fmt.Errorf("clickhouse backup creating operation is %q status long time", last.Status)
			}
		}

		return nil
	}

	if err := backoff.Retry(op, bo); err != nil {
		b.Status.Phase = "CreateFailed"
		b.Status.Error = err.Error()
	}

	return rc.Status().Update(ctx, b)
}

func uploadClickHouseBackup(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
	b.Status.Phase = "Uploading"
	if err := rc.Status().Update(ctx, b); err != nil {
		return fmt.Errorf("failed update clickhouse backup object: %w", err)
	}

	if _, err := clickhouse.UploadBackup(ctx, b); err != nil {
		b.Status.Phase = "UploadFailed"
		b.Status.Error = err.Error()
		if err := rc.Status().Update(ctx, b); err != nil {
			return err
		}

		return err
	}

	bo := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	op := func() error {
		rows, err := clickhouse.GetStatus(ctx, b)
		if err != nil {
			return fmt.Errorf("failed to get backups status: %w", err)
		}

		if len(rows) > 0 {
			last := rows[len(rows)-1]
			switch last.Status {
			case "error":
				b.Status.Phase = "UploadFailed"
				b.Status.Error = last.Error
				if err := rc.Status().Update(ctx, b); err != nil {
					return backoff.Permanent(err)
				}

				return backoff.Permanent(errors.New("clickhouse backup uploading failed"))
			case "success":
				b.Status.Phase = "Completed"
				return rc.Status().Update(ctx, b)
			default:
				return fmt.Errorf("clickhouse backup uploading operation is %q status long time", last.Status)
			}
		}

		return nil
	}

	if err := backoff.Retry(op, bo); err != nil {
		b.Status.Phase = "UploadFailed"
		b.Status.Error = err.Error()
	}

	return rc.Status().Update(ctx, b)
}
