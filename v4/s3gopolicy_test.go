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
	"github.com/stretchr/testify/require"
)

func TestCreatePolicies(t *testing.T) {
	s3gopolicy.NowTime = func() time.Time {
		return time.Date(2016, time.December, 10, 0, 0, 0, 0, time.UTC)
	}

	policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
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
	require.NoError(t, err)

	assert.Equal(t, "eyJleHBpcmF0aW9uIjoiMjAxNi0xMi0xMFQwMTowMDowMFpaIiwiY29uZGl0aW9ucyI6W3siYnVja2V0IjoidGVzdC5idWNrZXQifSx7ImtleSI6ImZpbGVzL2t5b2tvbWkvdGVzdC5tb3YifSx7IkNvbnRlbnQtVHlwZSI6InZpZGVvL3F1aWNrdGltZSJ9LFsiY29udGVudC1sZW5ndGgtcmFuZ2UiLDExMzM4MTU1OCwxMTMzODE1NThdLHsieC1hbXotY3JlZGVudGlhbCI6Ilx1MDAzY0FXU19BQ0NFU1NfS0VZX0lEXHUwMDNlLzIwMTYxMjEwL2FwLW5vcnRoZWFzdC0xL3MzL2F3czRfcmVxdWVzdCJ9LHsieC1hbXotYWxnb3JpdGhtIjoiQVdTNC1ITUFDLVNIQTI1NiJ9LHsieC1hbXotZGF0ZSI6IjIwMTYxMjEwVDAwMDAwMFoifSx7IngtYW16LW1ldGEtZmlsZU5hbWUiOiJ0ZXN0Lm1vdiJ9XX0=",
		policies.Form["Policy"])
	assert.Equal(t, "21678aaeddd0c8f3082c891321c18d89e4007b0ca20f2909268a87f0bf2522e9",
		policies.Form["X-Amz-Signature"])
}

func ExampleCreatePolicies() {
	policies, err := s3gopolicy.CreatePolicies(s3gopolicy.AWSCredentials{
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

	if err != nil {
		log.Fatalln(err)
	}

	if err := testUpload(policies.URL, "files/kyokomi/test.mov", policies); err != nil {
		log.Fatalln(err)
	}
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
