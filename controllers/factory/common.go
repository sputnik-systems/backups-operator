package factory

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/robfig/cron/v3"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PhaseStarted      = "Started"
	PhaseFailed       = "Failed"
	PhaseCompleted    = "Completed"
	PhaseCreating     = "Creating"
	PhaseCreateFailed = "CreateFailed"
	PhaseUploading    = "Uploading"
	PhaseUploadFailed = "UploadFailed"
)

func ScheduleTask(c *cron.Cron, l logr.Logger, schedule string, id int, f func()) (cron.EntryID, error) {
	if id != 0 {
		eId := cron.EntryID(id)

		for _, entry := range c.Entries() {
			if entry.ID == eId {
				l.V(4).Info("removing task schedule", "schedule", schedule, "id", strconv.Itoa(id))

				c.Remove(eId)
			}
		}
	}

	l.V(4).Info("scheduling task", "schedule", schedule)

	return c.AddFunc(schedule, f)
}

func getCredentials(ctx context.Context, rc client.Client, secrets []string, ns string) (map[string]string, error) {
	creds := make(map[string]string)

	for _, name := range secrets {
		s := &v1.Secret{}
		n := types.NamespacedName{Namespace: ns, Name: name}
		err := rc.Get(ctx, n, s)
		if err != nil {
			return nil, err
		}

		for k, v := range s.Data {
			creds[k] = string(v)
		}
	}

	return creds, nil
}

func getFQDN(rawUrl, ns string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	hostname := u.Hostname()
	if len(strings.Split(hostname, ".")) < 2 {
		host := []string{hostname, ns, "svc"}
		u.Host = strings.Join(host, ".") + ":" + u.Port()
	}

	return u.String(), nil
}

func getUrlWithIP(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %s", err)
	}

	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return "", fmt.Errorf("failed to lookup ips: %s", err)
	}

	u.Host = ips[0].String() + ":" + u.Port()

	return u.String(), nil
}

func getHostname(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %s", err)
	}

	return u.Hostname(), nil
}
