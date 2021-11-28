package factory

import (
	"context"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
	"github.com/sputnik-systems/backups-operator/internal/clickhouse"
)

func UpdateClickHouseBackupStatusApiInfo(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
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

func CreateClickHouseBackup(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
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
				return fmt.Errorf("clickhouse backup creating operation is %q", last.Status)
			}
		}

		return nil
	}

	return backoff.Retry(op, bo)
}

func UploadClickHouseBackup(ctx context.Context, rc client.Client, b *backupsv1alpha1.ClickHouseBackup) error {
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
				return fmt.Errorf("clickhouse backup uploading operation is %q", last.Status)
			}
		}

		return nil
	}

	return backoff.Retry(op, bo)
}
