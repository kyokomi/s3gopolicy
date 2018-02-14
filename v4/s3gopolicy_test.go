package s3gopolicy_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kyokomi/s3gopolicy/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreatePolicies(t *testing.T) {
	as := assert.New(t)
	_ = as

	s3gopolicy.NowTime = func() time.Time {
		return time.Date(2016, time.December, 10, 0, 0, 0, 0, time.Local)
	}

	policies, _ := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
		Region:         "ap-northeast-1",
		AWSAccessKeyID: "<AWS_ACCESS_KEY_ID>",
		AWSSecretKeyID: "<AWS_SECRET_KEY_ID>",
	}, s3gopolicy.UploadConfig{
		UploadURL:   "https://s3-ap-northeast-1.amazonaws.com/test.bucket",
		BucketName:  "test.bucket",
		ObjectKey:   "files/kyokomi/test.mov",
		ContentType: "video/quicktime",
		FileSize:    113381558,
		MetaData: map[string]string{
			"x-amz-meta-fileName": "test.mov",
		},
	})

	as.Equal(policies.Form["Policy"], "eyJleHBpcmF0aW9uIjoiMjAxNi0xMi0xMFQwMTowMDowMForMDk6MDAiLCJjb25kaXRpb25zIjpbeyJidWNrZXQiOiJ0ZXN0LmJ1Y2tldCJ9LHsia2V5IjoiZmlsZXMva3lva29taS90ZXN0Lm1vdiJ9LHsiQ29udGVudC1UeXBlIjoidmlkZW8vcXVpY2t0aW1lIn0sWyJjb250ZW50LWxlbmd0aC1yYW5nZSIsMTEzMzgxNTU4LDExMzM4MTU1OF0seyJ4LWFtei1jcmVkZW50aWFsIjoiXHUwMDNjQVdTX0FDQ0VTU19LRVlfSURcdTAwM2UvMjAxNjEyMDkvYXAtbm9ydGhlYXN0LTEvczMvYXdzNF9yZXF1ZXN0In0seyJ4LWFtei1hbGdvcml0aG0iOiJBV1M0LUhNQUMtU0hBMjU2In0seyJ4LWFtei1kYXRlIjoiMjAxNjEyMTBUMDAwMDAwWiJ9LHsieC1hbXotbWV0YS1maWxlTmFtZSI6InRlc3QubW92In1dfQ==")
	as.Equal(policies.Form["X-Amz-Signature"], "dea7915cd0c356e393018195d0876b46f9c7f91d9daa4ff3ba90b12cc6e6d6c2")
}

func testUpload(url, file string, policies s3gopolicy.UploadPolicies) (err error) {
	// Add your image file
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	fileInfo, _ := f.Stat()
	log.Println(fileInfo.Size())

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for k, v := range policies.Form {
		if err := w.WriteField(k, v); err != nil {
			return err
		}
	}

	fw, err := w.CreateFormFile("file", file)
	if err != nil {
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		return
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	if err := w.Close(); err != nil {
		return err
	}

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	err = fmt.Errorf("status code: %s", res.Status)
	data, _ := ioutil.ReadAll(res.Body)
	log.Printf("response:%s\n", string(data))
	return
}
