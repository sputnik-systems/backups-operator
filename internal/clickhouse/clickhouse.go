package clickhouse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AlexAkulov/clickhouse-backup/pkg/server"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
)

type Backup struct {
	Name           string `json:"name"`
	Created        string `json:"created"`
	Size           int64  `json:"size,omitempty"`
	Location       string `json:"location"`
	RequiredBackup string `json:"required"`
	Desc           string `json:"desc"`
}

func CreateBackup(ctx context.Context, b *backupsv1alpha1.ClickHouseBackup) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, b.Status.Api.Address+"/backup/create", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup creation request: %s", err)
	}

	q := req.URL.Query()
	for key, value := range b.Spec.CreateParams {
		q.Add(key, value)
	}
	q.Add("name", b.Name)
	req.URL.RawQuery = q.Encode()

	return http.DefaultClient.Do(req)
}

func UploadBackup(ctx context.Context, b *backupsv1alpha1.ClickHouseBackup) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, b.Status.Api.Address+"/backup/upload/"+b.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup uploading request: %s", err)
	}

	q := req.URL.Query()
	for key, value := range b.Spec.CreateParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	return http.DefaultClient.Do(req)
}

func DeleteBackup(ctx context.Context, b *backupsv1alpha1.ClickHouseBackup) (*http.Response, error) {
	backups, err := listBackups(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %s", err)
	}

	for _, backup := range backups {
		switch backup.Location {
		case "local":
			if resp, err := http.Post(b.Spec.ApiAddress+"/backup/delete/local/"+b.Name, "application/json", nil); err != nil {
				return resp, fmt.Errorf("failed to delete local backup: %s", err)
			}
		case "remote":
			if resp, err := http.Post(b.Spec.ApiAddress+"/backup/delete/remote/"+b.Name, "application/json", nil); err != nil {
				return resp, fmt.Errorf("failed to delete remote backup: %s", err)
			}
		}
	}

	return nil, nil
}

func GetStatus(ctx context.Context, b *backupsv1alpha1.ClickHouseBackup) ([]server.ActionRow, error) {
	resp, err := http.Get(b.Status.Api.Address + "/backup/status")
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %s", err)
	}

	scanner := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()

	rows := make([]server.ActionRow, 0)
	for scanner.Scan() {
		var row server.ActionRow
		err := json.Unmarshal(scanner.Bytes(), &row)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal action row: %s", err)
		}

		if strings.Contains(row.Command, b.Name) {
			rows = append(rows, row)
		}
	}

	return rows, nil
}

func listBackups(ctx context.Context, b *backupsv1alpha1.ClickHouseBackup) ([]Backup, error) {
	resp, err := http.Get(b.Spec.ApiAddress + "/backup/list")
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %s", err)
	}

	scanner := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()

	backups := make([]Backup, 0)
	for scanner.Scan() {
		var backup Backup
		err := json.Unmarshal(scanner.Bytes(), &backup)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal backup: %s", err)
		}

		if backup.Name == b.Name {
			backups = append(backups, backup)
		}
	}

	return backups, nil
}
