package s3gopolicy_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyokomi/s3gopolicy/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreatePolicies(t *testing.T) {
	as := assert.New(t)

	s3gopolicy.NowTime = func() time.Time {
		return time.Date(2016, time.December, 10, 0, 0, 0, 0, time.UTC)
	}

	policies, _ := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
		AWSAccessKeyID: "AWS_ACCESS_KEY_ID",
		AWSSecretKeyID: "AWS_SECRET_KEY_ID",
	}, s3gopolicy.UploadConfig{
		BucketName:  "hogehogefugafuga.amazonaws.com",
		ObjectKey:   "files/image1.png",
		ContentType: "image/png",
		FileSize:    2000,
	})

	fmt.Printf("%#v\n", policies)

	as.Equal("eyJleHBpcmF0aW9uIjoiMjAxNi0xMi0xMFQwMTowMDowMC4wMDBaIiwiY29uZGl0aW9ucyI6W3siYnVja2V0IjoiaG9nZWhvZ2VmdWdhZnVnYS5hbWF6b25hd3MuY29tIn0seyJrZXkiOiJmaWxlcy9pbWFnZTEucG5nIn0seyJDb250ZW50LVR5cGUiOiJpbWFnZS9wbmcifSxbImNvbnRlbnQtbGVuZ3RoLXJhbmdlIiwyMDAwLDIwMDBdXX0=",
		policies.Form.Policy)
	as.Equal("ToFz/ggBRhVyYBRjUUz718HSLA8=",
		policies.Form.Signature)
}

// ローカルタイムがUTC以外でも、UTCで生成した場合と同じ結果になることを確認する
func TestCreatePoliciesNonUTC(t *testing.T) {
	as := assert.New(t)

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	s3gopolicy.NowTime = func() time.Time {
		return time.Date(2016, time.December, 10, 9, 0, 0, 0, jst) // UTCの2016-12-10T00:00:00Zと同時刻
	}

	policies, _ := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
		AWSAccessKeyID: "AWS_ACCESS_KEY_ID",
		AWSSecretKeyID: "AWS_SECRET_KEY_ID",
	}, s3gopolicy.UploadConfig{
		BucketName:  "hogehogefugafuga.amazonaws.com",
		ObjectKey:   "files/image1.png",
		ContentType: "image/png",
		FileSize:    2000,
	})

	as.Equal("eyJleHBpcmF0aW9uIjoiMjAxNi0xMi0xMFQwMTowMDowMC4wMDBaIiwiY29uZGl0aW9ucyI6W3siYnVja2V0IjoiaG9nZWhvZ2VmdWdhZnVnYS5hbWF6b25hd3MuY29tIn0seyJrZXkiOiJmaWxlcy9pbWFnZTEucG5nIn0seyJDb250ZW50LVR5cGUiOiJpbWFnZS9wbmcifSxbImNvbnRlbnQtbGVuZ3RoLXJhbmdlIiwyMDAwLDIwMDBdXX0=",
		policies.Form.Policy)
	as.Equal("ToFz/ggBRhVyYBRjUUz718HSLA8=",
		policies.Form.Signature)
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
