s3gopolicy
=========================================

[![CircleCI](https://circleci.com/gh/kyokomi/s3gopolicy/tree/master.svg?style=svg)](https://circleci.com/gh/kyokomi/s3gopolicy/tree/master)

Authenticating Requests in Browser-Based Uploads Using POST (AWS Signature Version 2 or 4) for golang

## AWS Signature Version v4
https://docs.aws.amazon.com/ja_jp/AmazonS3/latest/API/sigv4-authentication-HTTPPOST.html

### Install

```shell
go get github.com/kyokomi/s3gopolicy/v4
```

### Usage

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/kyokomi/s3gopolicy/v4"
)

func main() {
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

	if err := openFileUpload(policies.URL, "./test.mov", policies); err != nil {
		log.Fatal(err)
	}
}

func openFileUpload(url, file string, policies s3gopolicy.UploadPolicies) (err error) {
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
```

## v2
http://docs.aws.amazon.com/ja_jp/AmazonS3/latest/dev/UsingHTTPPOST.html

### Install

```shell
go get github.com/kyokomi/s3gopolicy/v2
```
