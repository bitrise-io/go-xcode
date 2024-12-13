package _integration_tests

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

// BucketAccessor ...
type BucketAccessor struct {
	jwtConfig    *jwt.Config
	bucket       string
	objectExpiry time.Duration
}

// NewBucketAccessor ...
func NewBucketAccessor(serviceAccountJSONContent, bucket string) (*BucketAccessor, error) {
	conf, err := google.JWTConfigFromJSON([]byte(serviceAccountJSONContent))
	if err != nil {
		return nil, err
	}

	return &BucketAccessor{
		jwtConfig:    conf,
		bucket:       bucket,
		objectExpiry: 1 * time.Hour,
	}, nil
}

// GetExpiringURL ...
func (a BucketAccessor) GetExpiringURL(originalURL string) (string, error) {
	artifactPath := strings.TrimPrefix(strings.TrimPrefix(originalURL, fmt.Sprintf("https://storage.googleapis.com/%s/", a.bucket)), fmt.Sprintf("https://storage.cloud.google.com/%s/", a.bucket))
	opts := &storage.SignedURLOptions{
		Method:         http.MethodGet,
		GoogleAccessID: a.jwtConfig.Email,
		PrivateKey:     a.jwtConfig.PrivateKey,
		Expires:        time.Now().Add(a.objectExpiry),
	}

	return storage.SignedURL(a.bucket, artifactPath, opts)
}
