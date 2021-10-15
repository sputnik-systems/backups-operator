package factory

import (
	"context"
	"net/url"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetCredentials(ctx context.Context, rc client.Client, secrets []string, ns string) (map[string]string, error) {
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

func GetFQDN(rawUrl, ns string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	hostname := u.Hostname()
	port := u.Port()
	if len(strings.Split(hostname, ".")) < 2 {
		host := []string{hostname, ns, "svc"}
		u.Host = strings.Join(host, ".") + ":" + port
	}

	return u.String(), nil
}
