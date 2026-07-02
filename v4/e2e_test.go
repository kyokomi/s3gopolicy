//go:build e2e

package s3gopolicy_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyokomi/s3gopolicy/v4"
	"github.com/stretchr/testify/require"
)

var e2eHTTPClient = &http.Client{Timeout: 10 * time.Second}

func e2eEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func e2eCreatePolicies(t *testing.T, content []byte) s3gopolicy.UploadPolicies {
	t.Helper()

	endpoint := strings.TrimSuffix(e2eEnv("S3GOPOLICY_E2E_ENDPOINT", "http://localhost:9000"), "/")
	bucket := e2eEnv("S3GOPOLICY_E2E_BUCKET", "e2e-bucket")

	policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
		Region:         "us-east-1",
		AWSAccessKeyID: e2eEnv("S3GOPOLICY_E2E_ACCESS_KEY", "minioadmin"),
		AWSSecretKeyID: e2eEnv("S3GOPOLICY_E2E_SECRET_KEY", "minioadmin"),
	}, s3gopolicy.UploadConfig{
		UploadURL:   endpoint + "/" + bucket,
		BucketName:  bucket,
		ObjectKey:   "e2e/test.txt",
		ContentType: "text/plain",
		FileSize:    int64(len(content)),
		MetaData: map[string]string{
			"x-amz-meta-fileName": "test.txt",
		},
	})
	require.NoError(t, err)
	return policies
}

func e2eBuildForm(t *testing.T, fields map[string]string, content []byte) (io.Reader, string) {
	t.Helper()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	fw, err := w.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return &b, w.FormDataContentType()
}

func e2ePostForm(t *testing.T, url string, fields map[string]string, content []byte) (int, string) {
	t.Helper()

	body, contentType := e2eBuildForm(t, fields, content)
	req, err := http.NewRequest(http.MethodPost, url, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)

	res, err := e2eHTTPClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, string(resBody)
}

func TestE2EUpload(t *testing.T) {
	content := bytes.Repeat([]byte("a"), 1024)
	policies := e2eCreatePolicies(t, content)

	status, body := e2ePostForm(t, policies.URL, policies.Form, content)
	require.Equal(t, http.StatusNoContent, status, "upload should succeed: %s", body)
}

func TestE2EUploadRejectsTamperedSignature(t *testing.T) {
	content := bytes.Repeat([]byte("a"), 1024)
	policies := e2eCreatePolicies(t, content)
	policies.Form["X-Amz-Signature"] = "0000000000000000000000000000000000000000000000000000000000000000"

	status, body := e2ePostForm(t, policies.URL, policies.Form, content)
	require.Equal(t, http.StatusForbidden, status, "tampered signature should be rejected: %s", body)
}

func TestE2EUploadRejectsPolicyMismatch(t *testing.T) {
	content := bytes.Repeat([]byte("a"), 1024)
	policies := e2eCreatePolicies(t, content)
	tooLarge := bytes.Repeat([]byte("a"), 2048)

	status, body := e2ePostForm(t, policies.URL, policies.Form, tooLarge)
	require.Equal(t, http.StatusBadRequest, status, "policy mismatch should be rejected: %s", body)
}
