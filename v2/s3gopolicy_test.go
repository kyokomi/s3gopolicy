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
		return time.Date(2016, time.December, 10, 0, 0, 0, 0, time.Local)
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

	as.Equal(policies.Form.Policy, "eyJleHBpcmF0aW9uIjoiMjAxNi0xMi0xMFQwMTowMDowMForMDk6MDAiLCJjb25kaXRpb25zIjpbeyJidWNrZXQiOiJob2dlaG9nZWZ1Z2FmdWdhLmFtYXpvbmF3cy5jb20ifSx7ImtleSI6ImZpbGVzL2ltYWdlMS5wbmcifSx7IkNvbnRlbnQtVHlwZSI6ImltYWdlL3BuZyJ9LFsiY29udGVudC1sZW5ndGgtcmFuZ2UiLDIwMDAsMjAwMF1dfQ==")
	as.Equal(policies.Form.Signature, "W41MfAMOAdhiweHec9so/xdO5mo=")
}
