//go:build e2e

package s3gopolicy_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"github.com/kyokomi/s3gopolicy/v2"
	"github.com/stretchr/testify/require"
)

func e2eEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func e2eCreatePolicies(t *testing.T, content []byte) s3gopolicy.UploadPolicies {
	endpoint := e2eEnv("S3GOPOLICY_E2E_ENDPOINT", "http://localhost:9000")
	bucket := e2eEnv("S3GOPOLICY_E2E_BUCKET", "e2e-bucket")

	policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
		AWSAccessKeyID: e2eEnv("S3GOPOLICY_E2E_ACCESS_KEY", "minioadmin"),
		AWSSecretKeyID: e2eEnv("S3GOPOLICY_E2E_SECRET_KEY", "minioadmin"),
	}, s3gopolicy.UploadConfig{
		UploadURL:   endpoint + "/" + bucket,
		BucketName:  bucket,
		ObjectKey:   "e2e/test-v2.txt",
		ContentType: "text/plain",
		FileSize:    int64(len(content)),
	})
	require.NoError(t, err)
	return policies
}

func e2ePostForm(t *testing.T, url string, form s3gopolicy.PoliciesForm, content []byte) (int, string) {
	fields := map[string]string{
		"AWSAccessKeyId": form.AWSAccessKeyID,
		"key":            form.ObjectKey,
		"Content-Type":   form.ContentType,
		"policy":         form.Policy,
		"signature":      form.Signature,
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	fw, err := w.CreateFormFile("file", "test-v2.txt")
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, url, &b)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, string(body)
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
	policies.Form.Signature = "AAAAAAAAAAAAAAAAAAAAAAAAAAA="

	status, body := e2ePostForm(t, policies.URL, policies.Form, content)
	require.Equal(t, http.StatusForbidden, status, "tampered signature should be rejected: %s", body)
}
