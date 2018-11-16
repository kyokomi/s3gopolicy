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

	as.Equal("eyJleHBpcmF0aW9uIjoiMjAxNi0xMi0xMFQwMTowMDowMFpaIiwiY29uZGl0aW9ucyI6W3siYnVja2V0IjoiaG9nZWhvZ2VmdWdhZnVnYS5hbWF6b25hd3MuY29tIn0seyJrZXkiOiJmaWxlcy9pbWFnZTEucG5nIn0seyJDb250ZW50LVR5cGUiOiJpbWFnZS9wbmcifSxbImNvbnRlbnQtbGVuZ3RoLXJhbmdlIiwyMDAwLDIwMDBdXX0=",
		policies.Form.Policy)
	as.Equal("FPI4mtudW6IZjj05ZsOWvug3TZA=",
		policies.Form.Signature)
}
