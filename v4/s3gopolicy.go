package s3gopolicy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// AWSCredentials Amazon Credentials
type AWSCredentials struct {
	Region         string
	AWSSecretKeyID string
	AWSAccessKeyID string
}

// UploadConfig generate policies from config
type UploadConfig struct {
	// Required
	BucketName  string
	ObjectKey   string
	ContentType string
	FileSize    int64
	// Optional
	UploadURL  string
	Expiration time.Duration
	MetaData   map[string]string
}

// UploadPolicies Amazon s3 upload policies
type UploadPolicies struct {
	URL  string
	Form map[string]string
}

// PoliciesForm is upload policies formData
type PoliciesForm struct {
	ObjectKey   string `json:"key"`
	ContentType string `json:"Content-Type"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	ACL         string `json:"acl"`
}

// PolicyJSON is policy rule
type PolicyJSON struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

const (
	expirationTimeFormat     = "2006-01-02T15:04:05ZZ07:00"
	amzDateISO8601TimeFormat = "20060102T150405Z"
	shortTimeFormat          = "20060102"
	algorithm                = "AWS4-HMAC-SHA256"
	serviceName              = "s3"

	defaultUploadURLFormat = "http://%s.s3.amazonaws.com/" // <bucketName>
	defaultExpirationHour  = 1 * time.Hour
)

// NowTime mockable time.Now()
var NowTime = func() time.Time {
	return time.Now()
}

// CreatePolicies create amazon s3 to upload policies return
// https://docs.aws.amazon.com/ja_jp/AmazonS3/latest/API/sigv4-authentication-HTTPPOST.html
func CreatePolicies(awsCredentials AWSCredentials, uploadConfig UploadConfig) (UploadPolicies, error) {
	nowTime := NowTime()
	credential := awsCredentials.buildCredential(nowTime)
	data, err := buildSign(nowTime, credential, uploadConfig)
	if err != nil {
		return UploadPolicies{}, err
	}
	// 1. StringToSign
	policy := base64.StdEncoding.EncodeToString(data)
	// 2. Signing Key
	hash := hmac.New(sha256.New, buildSignature(nowTime, awsCredentials.AWSSecretKeyID, awsCredentials.Region, serviceName))
	hash.Write([]byte(policy))
	// 3. Signature
	signature := hex.EncodeToString(hash.Sum(nil))

	uploadURL := uploadConfig.UploadURL
	if uploadURL == "" {
		uploadURL = fmt.Sprintf(defaultUploadURLFormat, uploadConfig.BucketName)
	}
	form := map[string]string{
		"key":              uploadConfig.ObjectKey,
		"Content-Type":     uploadConfig.ContentType,
		"X-Amz-Credential": credential,
		"X-Amz-Algorithm":  algorithm,
		"X-Amz-Date":       nowTime.Format(amzDateISO8601TimeFormat),
		"Policy":           policy,
		"X-Amz-Signature":  signature,
	}
	for k, v := range uploadConfig.MetaData {
		form[k] = v
	}
	return UploadPolicies{
		URL:  uploadURL,
		Form: form,
	}, nil
}

func buildSign(nowTime time.Time, credential string, uploadConfig UploadConfig) ([]byte, error) {
	conditions := []interface{}{
		map[string]string{"bucket": uploadConfig.BucketName},
		map[string]string{"key": uploadConfig.ObjectKey},
		map[string]string{"Content-Type": uploadConfig.ContentType},
		[]interface{}{"content-length-range", uploadConfig.FileSize, uploadConfig.FileSize},
		map[string]string{"x-amz-credential": credential},
		map[string]string{"x-amz-algorithm": algorithm},
		map[string]string{"x-amz-date": nowTime.Format(amzDateISO8601TimeFormat)},
	}
	for k, v := range uploadConfig.MetaData {
		conditions = append(conditions, map[string]string{k: v})
	}

	expiration := defaultExpirationHour
	if uploadConfig.Expiration > 0 {
		expiration = uploadConfig.Expiration
	}

	return json.Marshal(&PolicyJSON{
		Expiration: nowTime.Add(expiration).Format(expirationTimeFormat),
		Conditions: conditions,
	})
}

func (c AWSCredentials) buildCredential(nowTime time.Time) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", c.AWSAccessKeyID, nowTime.UTC().Format(shortTimeFormat), c.Region, serviceName, "aws4_request")
}

func buildSignature(nowTime time.Time, secretAccessKey string, regionName string, serviceName string) []byte {
	shortTime := nowTime.UTC().Format(shortTimeFormat)

	date := makeHMac([]byte("AWS4"+secretAccessKey), []byte(shortTime))
	region := makeHMac(date, []byte(regionName))
	service := makeHMac(region, []byte(serviceName))
	credentials := makeHMac(service, []byte("aws4_request"))
	return credentials
}

func makeHMac(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}
