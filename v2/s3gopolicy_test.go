package s3gopolicy_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/kyokomi/s3gopolicy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockNowTime(t *testing.T, tm time.Time) {
	orig := s3gopolicy.NowTime
	s3gopolicy.NowTime = func() time.Time { return tm }
	t.Cleanup(func() { s3gopolicy.NowTime = orig })
}

func TestCreatePolicies(t *testing.T) {
	as := assert.New(t)

	mockNowTime(t, time.Date(2016, time.December, 10, 0, 0, 0, 0, time.UTC))

	policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
		AWSAccessKeyID: "AWS_ACCESS_KEY_ID",
		AWSSecretKeyID: "AWS_SECRET_KEY_ID",
	}, s3gopolicy.UploadConfig{
		BucketName:  "hogehogefugafuga.amazonaws.com",
		ObjectKey:   "files/image1.png",
		ContentType: "image/png",
		FileSize:    2000,
	})
	require.NoError(t, err)

	as.Equal("eyJleHBpcmF0aW9uIjoiMjAxNi0xMi0xMFQwMTowMDowMC4wMDBaIiwiY29uZGl0aW9ucyI6W3siYnVja2V0IjoiaG9nZWhvZ2VmdWdhZnVnYS5hbWF6b25hd3MuY29tIn0seyJrZXkiOiJmaWxlcy9pbWFnZTEucG5nIn0seyJDb250ZW50LVR5cGUiOiJpbWFnZS9wbmcifSxbImNvbnRlbnQtbGVuZ3RoLXJhbmdlIiwyMDAwLDIwMDBdXX0=",
		policies.Form.Policy)
	as.Equal("ToFz/ggBRhVyYBRjUUz718HSLA8=",
		policies.Form.Signature)

	policyJSON, err := base64.StdEncoding.DecodeString(policies.Form.Policy)
	require.NoError(t, err)
	var policy struct {
		Expiration string `json:"expiration"`
	}
	require.NoError(t, json.Unmarshal(policyJSON, &policy))
	as.Equal("2016-12-10T01:00:00.000Z", policy.Expiration)
}

// NowTimeがUTC以外のタイムゾーンでも、UTCと同じポリシー・署名になることを確認する
func TestCreatePoliciesNonUTC(t *testing.T) {
	utcTime := time.Date(2016, time.December, 10, 0, 0, 0, 0, time.UTC)
	jstTime := utcTime.In(time.FixedZone("Asia/Tokyo", 9*60*60))

	createPolicies := func(tm time.Time) s3gopolicy.UploadPolicies {
		mockNowTime(t, tm)
		policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
			AWSAccessKeyID: "AWS_ACCESS_KEY_ID",
			AWSSecretKeyID: "AWS_SECRET_KEY_ID",
		}, s3gopolicy.UploadConfig{
			BucketName:  "hogehogefugafuga.amazonaws.com",
			ObjectKey:   "files/image1.png",
			ContentType: "image/png",
			FileSize:    2000,
		})
		require.NoError(t, err)
		return policies
	}

	utcPolicies := createPolicies(utcTime)
	jstPolicies := createPolicies(jstTime)

	assert.Equal(t, utcPolicies.Form.Policy, jstPolicies.Form.Policy)
	assert.Equal(t, utcPolicies.Form.Signature, jstPolicies.Form.Signature)
}

func TestCreatePoliciesWithMetaData(t *testing.T) {
	as := assert.New(t)

	mockNowTime(t, time.Date(2016, time.December, 10, 0, 0, 0, 0, time.UTC))

	metaData := []map[string]string{
		{"x-amz-meta-fileName": "image1.png"},
	}
	policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
		AWSAccessKeyID: "AWS_ACCESS_KEY_ID",
		AWSSecretKeyID: "AWS_SECRET_KEY_ID",
	}, s3gopolicy.UploadConfig{
		BucketName:  "examplebucket",
		ObjectKey:   "files/image1.png",
		ContentType: "image/png",
		FileSize:    2000,
		MetaData:    metaData,
	})
	require.NoError(t, err)

	as.Equal(metaData, policies.MetaData)

	policyJSON, err := base64.StdEncoding.DecodeString(policies.Form.Policy)
	require.NoError(t, err)
	var policy struct {
		Conditions []any `json:"conditions"`
	}
	require.NoError(t, json.Unmarshal(policyJSON, &policy))
	as.Contains(policy.Conditions, map[string]any{"x-amz-meta-fileName": "image1.png"})
}

func TestCreatePoliciesDefaultUploadURL(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		wantURL    string
	}{
		{"bucket without dots uses https", "examplebucket", "https://examplebucket.s3.amazonaws.com/"},
		{"bucket with dots keeps http", "test.bucket", "http://test.bucket.s3.amazonaws.com/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
				AWSAccessKeyID: "AWS_ACCESS_KEY_ID",
				AWSSecretKeyID: "AWS_SECRET_KEY_ID",
			}, s3gopolicy.UploadConfig{
				BucketName:  tt.bucketName,
				ObjectKey:   "files/image1.png",
				ContentType: "image/png",
				FileSize:    2000,
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.wantURL, policies.URL)
		})
	}
}
