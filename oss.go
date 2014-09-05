package aliyun

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const Version = "developing..."

const (
	L_QingDao          = "oss-cn-qingdao.aliyuncs.com"
	L_QingDaoInternal  = "oss-cn-qingdao-internal.aliyuncs.com"
	L_HangZhou         = "oss-cn-hangzhou.aliyuncs.com"
	L_HangZhouInternal = "oss-cn-hangzhou-internal.aliyuncs.com"
	L_Beijing          = "oss-cn-beijing.aliyuncs.com"
	L_BeijingInternal  = "oss-cn-beijing-internal.aliyuncs.com"
	ACL_Private        = "private"
	ACL_Public_RDONLY  = "public-read"
	ACL_Public_RDRW    = "public-read-write"
)

var (
	c = []byte(":")
	n = []byte("\n")
)

type ErrorXML struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

type Client struct {
	accessID     string
	accessSecret string
	*http.Client
}

func New(accessID, accessSecret string) *Client {
	return &Client{accessID, accessSecret, new(http.Client)}
}

func (client *Client) PutBucket(bucketName, location, acl string) (bucket *Bucket, err error) {
	err = checkBucketName(bucketName)
	if err != nil {
		return
	}
	header := make(http.Header)
	header.Set("x-oss-acl", acl)
	err = client.do("PUT", bucketName, location, "", header, 200, nil)
	if err != nil {
		return
	}
	bucket = &Bucket{bucketName, location, client}
	return
}

func (client *Client) DeleteBucket(bucketName, location string) (err error) {
	err = checkBucketName(bucketName)
	if err != nil {
		return
	}
	return client.do("DELETE", bucketName, location, "", nil, 204, nil)
}

func (client *Client) do(method, bucket, location, object string, header http.Header, expectedCode int, body io.Reader) (err error) {
	object = strings.Trim(object, "/")
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s.%s/%s", bucket, location, object), body)
	if err != nil {
		return
	}

	if header == nil {
		header = make(http.Header)
	}
	header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	if _, ok := header["User-Agent"]; !ok {
		header.Set("User-Agent", "oss-golang")
	}
	client.makeSignature(method, bucket, object, header)
	req.Header = header

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == expectedCode {
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// fmt.Println(string(b))

	var ex ErrorXML
	if xml.Unmarshal(b, &ex) == nil {
		err = fmt.Errorf("%d %s: %s", resp.StatusCode, ex.Code, ex.Message)
		return
	}
	return fmt.Errorf("%s", resp.Status)
}

func (client *Client) makeSignature(method, bucket, object string, header http.Header) {
	hmac := hmac.New(sha1.New, []byte(client.accessSecret))
	hmac.Write([]byte(method))
	hmac.Write(n)

	hmac.Write([]byte(header.Get("Content-Md5")))
	hmac.Write(n)

	hmac.Write([]byte(header.Get("Content-Type")))
	hmac.Write(n)

	hmac.Write([]byte(header.Get("Date")))
	hmac.Write(n)

	for key, value := range header {
		//todo sort map..
		key = strings.ToLower(key)
		if strings.HasPrefix(key, "x-oss-") {
			hmac.Write([]byte(key))
			hmac.Write(c)
			hmac.Write([]byte(strings.Join(value, ",")))
			hmac.Write(n)
		}
	}

	hmac.Write([]byte(fmt.Sprintf("/%s/%s", bucket, object)))

	header.Set("Authorization", fmt.Sprintf("OSS %s:%s", client.accessID, base64.StdEncoding.EncodeToString(hmac.Sum(nil))))
}

type Bucket struct {
	name     string
	location string
	*Client
}

func (bucket *Bucket) PutObject(objectName, contentType string, content io.ReadSeeker, headers map[string]string) (err error) {
	if content == nil {
		return
	}
	length, err := content.Seek(0, os.SEEK_END)
	if err != nil {
		return
	}
	_, err = content.Seek(0, os.SEEK_SET)
	if err != nil {
		return
	}
	header := make(http.Header)
	header.Set("Content-Type", contentType)
	header.Set("Content-Length", fmt.Sprintf("%d", length))
	for key, val := range headers {
		header.Set(key, val)
	}
	return bucket.do("PUT", bucket.name, bucket.location, objectName, header, 200, content)
}

func (bucket *Bucket) DeleteObject(objectName string) error {
	return bucket.do("DELETE", bucket.name, bucket.location, objectName, nil, 204, nil)
}

func checkBucketName(name string) error {
	nl := len(name)
	if nl < 3 || nl > 63 {
		return fmt.Errorf("400 InvalidBucketName: Invalid bucket name '%s'", name)
	}
	var c byte
	for i := 0; i < nl; i++ {
		if c = name[i]; (c < '0' && c != '-') || (c > '9' && c < 'a') || c > 'z' {
			return fmt.Errorf("400 InvalidBucketName: Invalid bucket name '%s'", name)
		}
	}
	if name[0] == '-' {
		return fmt.Errorf("400 InvalidBucketName: Invalid bucket name '%s'", name)
	}
	return nil
}
