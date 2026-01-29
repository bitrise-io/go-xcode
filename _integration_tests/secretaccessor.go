package _integration_tests

import (
	"context"
	"fmt"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/v2/retry"
	"google.golang.org/api/option"
)

// SecretAccessor ...
type SecretAccessor struct {
	ctx       context.Context
	client    *secretmanager.Client
	projectID string
}

// NewSecretAccessor ...
func NewSecretAccessor(serviceAccountJSONContent, projectID string) (*SecretAccessor, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx, option.WithCredentialsJSON([]byte(serviceAccountJSONContent)))
	if err != nil {
		return nil, err
	}

	return &SecretAccessor{
		ctx:       ctx,
		client:    client,
		projectID: projectID,
	}, nil
}

// GetSecret ...
func (m SecretAccessor) GetSecret(key string) (string, error) {
	secretValue := ""
	if err := retry.Times(3).Wait(30 * time.Second).Try(func(attempt uint) error {
		if attempt > 0 {
			log.Warnf("%d attempt failed", attempt)
		}

		name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", m.projectID, key)
		req := &secretmanagerpb.AccessSecretVersionRequest{
			Name: name,
		}
		result, err := m.client.AccessSecretVersion(m.ctx, req)
		if err != nil {
			log.Warnf("%s", err)
			return err
		}

		secretValue = string(result.Payload.Data)
		return nil
	}); err != nil {
		return "", err
	}

	return secretValue, nil
}
