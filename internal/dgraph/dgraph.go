package dgraph

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/hasura/go-graphql-client"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sputnik-systems/backups-storage/s3"

	backupsv1alpha1 "github.com/sputnik-systems/backups-operator/api/v1alpha1"
)

type ExportOutput struct {
	Response struct {
		Message graphql.String
		Code    graphql.String
	}

	ExportedFiles []graphql.String
}

func Export(ctx context.Context, rc client.Client, bs *backupsv1alpha1.DgraphBackupSpec, creds map[string]string) (*ExportOutput, error) {
	type ExportInput struct {
		Format       graphql.String  `json:"format"`
		Namespace    graphql.Int     `json:"namespace"`
		Destination  graphql.String  `json:"destination"`
		AccessKey    graphql.String  `json:"accessKey"`
		SecretKey    graphql.String  `json:"secretKey"`
		SessionToken graphql.String  `json:"sessionToken"`
		Anonymous    graphql.Boolean `json:"anonymous"`
	}

	input := ExportInput{
		Format:      graphql.String(bs.Format),
		Namespace:   graphql.Int(bs.Namespace),
		Destination: graphql.String(bs.Destination),
		Anonymous:   graphql.Boolean(bs.Anonymous),
	}

	id, secret, token := parseCredentials(creds)
	input.AccessKey = graphql.String(id)
	input.SecretKey = graphql.String(secret)
	input.SessionToken = graphql.String(token)

	gqlVars := map[string]interface{}{
		"input": input,
	}

	var gqlMutation struct {
		ExportOutput `graphql:"export(input: $input)"`
	}

	client := graphql.NewClient(bs.AdminUrl, nil)
	err := client.Mutate(context.Background(), &gqlMutation, gqlVars)
	if err != nil {
		return nil, err
	}

	return &gqlMutation.ExportOutput, nil
}

func DeleteExport(ctx context.Context, b *backupsv1alpha1.DgraphBackup, creds map[string]string) error {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return err
	}

	u, err := url.Parse(b.Spec.Destination)
	if err != nil {
		return err
	}

	endpoint := u.Hostname()
	uri := strings.Split(u.RequestURI(), "/")

	id, secret, token := parseCredentials(creds)
	sess.Config.WithEndpoint(endpoint)
	sess.Config.WithRegion(b.Spec.Region)
	sess.Config.WithS3ForcePathStyle(true)
	sess.Config.WithCredentials(
		credentials.NewStaticCredentials(id, secret, token))

	bucket := uri[1]
	prefix := path.Join(uri[2:]...)
	backup := path.Dir(b.Status.ExportResponse.ExportedFiles[0])
	storage := s3.NewStorage(sess, bucket, prefix)
	return storage.Delete(backup)
}

func parseCredentials(creds map[string]string) (id, secret, token string) {
	if value, ok := creds["accessKey"]; ok {
		id = value
	}

	if value, ok := creds["secretKey"]; ok {
		secret = value
	}

	if value, ok := creds["sessionToken"]; ok {
		token = value
	}

	return
}
