/*
 * Copyright (c) 2018 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/logging"
	"github.com/spf13/viper"
)

// S3API defines the interface for S3 operations
type S3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// Client is a wrapper over the official aws s3 package
// only implements used functions
type Client struct {
	client       S3API
	presignAPI   *s3.PresignClient
	bucket       string
	folder       string
	endpointURL  string
	forcePathStyle bool
}

// NewClient ctor
func NewClient(prefix string, conf *viper.Viper) (*Client, error) {
	ctx := context.Background()
	region := conf.GetString(fmt.Sprintf("%s.region", prefix))
	accessKey := conf.GetString(fmt.Sprintf("%s.accessKey", prefix))
	secretAccessKey := conf.GetString(fmt.Sprintf("%s.secretAccessKey", prefix))
	endpoint := conf.GetString(fmt.Sprintf("%s.endpoint", prefix))

	// Create AWS config with credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretAccessKey,
			"",
		)),
		config.WithLogger(logging.NewStandardLogger(io.Discard)),
	)
	if err != nil {
		return nil, err
	}

	// Configure S3 client with options
	var s3Options []func(*s3.Options)
	forcePathStyle := true
	
	if endpoint != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = forcePathStyle
		})
	}

	svc := s3.NewFromConfig(cfg, s3Options...)
	presignClient := s3.NewPresignClient(svc)

	return &Client{
		client:         svc,
		presignAPI:     presignClient,
		bucket:         conf.GetString(fmt.Sprintf("%s.bucket", prefix)),
		folder:         conf.GetString(fmt.Sprintf("%s.folder", prefix)),
		endpointURL:    endpoint,
		forcePathStyle: forcePathStyle,
	}, nil
}

func streamToByte(stream io.ReadCloser) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

// GetObject gets an object from s3
func (c Client) GetObject(path string) ([]byte, error) {
	ctx := context.Background()
	splittedString := strings.SplitN(path, "/", 2)
	if len(splittedString) < 2 {
		return nil, fmt.Errorf("Invalid path")
	}
	bucket := splittedString[0]
	objKey := splittedString[1]
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objKey),
	}
	resp, err := c.client.GetObject(ctx, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return streamToByte(resp.Body), nil
}

// PutObject puts an object into s3
func (c Client) PutObject(ctx context.Context, path string, body *[]byte) error {
	b := bytes.NewReader(*body)
	params := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
		Body:   b,
	}
	_, err := c.client.PutObject(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

// MakePath concatenates folder with key
func (c Client) MakePath(k string) string {
	return fmt.Sprintf("%s/%s", c.folder, k)
}

// PutObjectRequest return a presigned url for uploading a file to s3
func (c Client) PutObjectRequest(ctx context.Context, key, acl string) (string, http.Header, error) {
	path := c.MakePath(key)
	params := &s3.PutObjectInput{
		ACL:    types.ObjectCannedACL(acl),
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
	}

	req, err := c.presignAPI.PresignPutObject(ctx, params, func(opts *s3.PresignOptions) {
		opts.Expires = 900 * time.Second
	})
	if err != nil {
		return "", nil, err
	}

	// Convert map to http.Header
	header := make(http.Header)
	for k, v := range req.SignedHeader {
		header[k] = v
	}

	return req.URL, header, nil
}

// DeleteObject deletes an object from s3
func (c Client) DeleteObject(ctx context.Context, key string) error {
	path := c.MakePath(key)
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
	}
	_, err := c.client.DeleteObject(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

// PutObjectInput puts an object into s3, if params.Bucket or params.Body
// are equal nil, they will be overwrite
func (c Client) PutObjectInput(ctx context.Context, params *s3.PutObjectInput, body *[]byte) error {
	b := bytes.NewReader(*body)
	if params.Bucket == nil {
		params.Bucket = aws.String(c.bucket)
	}
	if params.Body == nil {
		params.Body = b
	}
	_, err := c.client.PutObject(ctx, params)
	if err != nil {
		return err
	}
	return nil
}
