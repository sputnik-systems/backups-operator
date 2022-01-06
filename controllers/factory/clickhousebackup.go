package factory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
	"github.com/sputnik-systems/backups-operator/controllers/factory/finalize"
	"github.com/sputnik-systems/backups-operator/internal/clickhouse"
)

func ProccessClickHouseBackupObject(ctx context.Context, rc client.Client, l logr.Logger, b *backupsv1alpha1.ClickHouseBackup) error {
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
	}

	if err := createClickHouseBackup(ctx, rc, l, b); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if err := uploadClickHouseBackup(ctx, rc, l, b); err != nil {
		return fmt.Errorf("failed to upload backup: %w", err)
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

func createClickHouseBackup(ctx context.Context, rc client.Client, l logr.Logger, b *backupsv1alpha1.ClickHouseBackup) error {
	if b.Status.Phase == "Started" {
		if _, err := clickhouse.CreateBackup(ctx, b); err != nil {
			b.Status.Phase = "CreateFailed"
			b.Status.Error = err.Error()
			if err := rc.Status().Update(ctx, b); err != nil {
				return err
			}

			return err
		}

		b.Status.Phase = "Creating"
		if err := rc.Status().Update(ctx, b); err != nil {
			return fmt.Errorf("failed update clickhouse backup object: %w", err)
		}

		l.V(4).Info("started backup creation")
	}

	if b.Status.Phase == "Creating" {
		l.V(4).Info("checking backup creation")

		var bo backoff.BackOff
		bo, err := b.Spec.ExponentialBackOff.GetBackOff()
		if err != nil {
			return fmt.Errorf("failed to parse backoff settings: %w", err)
		}

		if bo, ok := bo.(*backoff.ExponentialBackOff); ok {
			if time.Since(b.CreationTimestamp.Time) > bo.MaxElapsedTime {
				b.Status.Phase = "CreateFailed"
				b.Status.Error = "backup creation timed out"

				return rc.Status().Update(ctx, b)
			}
		}

		op := func() error {
			rows, err := clickhouse.GetStatus(ctx, b)
			if err != nil {
				return fmt.Errorf("failed to get backups status: %w", err)
			}

			l.V(4).Info("backup creation progress", "rows", strconv.Itoa(len(rows)))

			if len(rows) > 0 {
				last := rows[len(rows)-1]

				l.V(4).Info("backup creation progress", "status", last.Status)

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

			return fmt.Errorf("clickhouse backup creating operation not found")
		}

		if err := backoff.Retry(op, backoff.WithContext(bo, ctx)); err != nil {
			b.Status.Phase = "CreateFailed"
			b.Status.Error = err.Error()
		}
	}

	return rc.Status().Update(ctx, b)
}

func uploadClickHouseBackup(ctx context.Context, rc client.Client, l logr.Logger, b *backupsv1alpha1.ClickHouseBackup) error {
	if b.Status.Phase == "Created" {
		if _, err := clickhouse.UploadBackup(ctx, b); err != nil {
			b.Status.Phase = "UploadFailed"
			b.Status.Error = err.Error()
			if err := rc.Status().Update(ctx, b); err != nil {
				return err
			}

			return err
		}

		b.Status.Phase = "Uploading"
		if err := rc.Status().Update(ctx, b); err != nil {
			return fmt.Errorf("failed update clickhouse backup object: %w", err)
		}

		l.V(4).Info("started backup uploading")
	}

	if b.Status.Phase == "Uploading" {
		l.V(4).Info("checking backup uploading")

		var bo backoff.BackOff
		bo, err := b.Spec.ExponentialBackOff.GetBackOff()
		if err != nil {
			return fmt.Errorf("failed to parse backoff settings: %w", err)
		}

		if bo, ok := bo.(*backoff.ExponentialBackOff); ok {
			if time.Since(b.CreationTimestamp.Time) > bo.MaxElapsedTime {
				b.Status.Phase = "UploadFailed"
				b.Status.Error = "backup creation timed out"

				return rc.Status().Update(ctx, b)
			}
		}

		op := func() error {
			rows, err := clickhouse.GetStatus(ctx, b)
			if err != nil {
				return fmt.Errorf("failed to get backups status: %w", err)
			}

			l.V(4).Info("backup uploading progress", "rows", strconv.Itoa(len(rows)))

			if len(rows) > 0 {
				last := rows[len(rows)-1]

				l.V(4).Info("backup uploading progress", "status", last.Status)

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

			return fmt.Errorf("clickhouse backup uploading operation not found")
		}

		if err := backoff.Retry(op, backoff.WithContext(bo, ctx)); err != nil {
			b.Status.Phase = "UploadFailed"
			b.Status.Error = err.Error()
		}
	}

	return rc.Status().Update(ctx, b)
}
