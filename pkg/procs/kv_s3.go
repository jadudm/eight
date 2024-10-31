package procs

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	aws_s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/philippgille/gokv/encoding"
	gokv_s3 "github.com/philippgille/gokv/s3"

	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/util"
)

type S3 struct {
	Bucket     env.Bucket
	Options    gokv_s3.Options
	s3_client  *s3.S3
	s3_session *session.Session
}

func NewKVS3(b env.Bucket) *S3 {
	options := gokv_s3.Options{
		BucketName:             b.Name,
		Region:                 b.Credentials.Region,
		CustomEndpoint:         b.Credentials.Endpoint,
		UsePathStyleAddressing: true,
		AWSaccessKeyID:         b.Credentials.AccessKeyId,
		AWSsecretAccessKey:     b.Credentials.SecretAccessKey,
		Codec:                  encoding.JSON,
	}
	s3 := S3{
		Bucket:  b,
		Options: options,
	}

	s3_c, s3_s := initClient(b)
	s3.s3_client = s3_c
	s3.s3_session = s3_s

	log.Println("s3 init")
	return &s3
}

func (obj *S3) Store(key string, value map[string]string) error {

	client, err := gokv_s3.NewClient(obj.Options)

	if err != nil {
		log.Println("s3 client")
		log.Fatal(err)
	}
	defer client.Close()

	err = client.Set(key, value)

	if err != nil {
		log.Println("s3 Set()")
		log.Fatal(err)
	}

	return nil
}

func (obj *S3) CreateBucket() {
	cparams := &aws_s3.CreateBucketInput{
		Bucket: &obj.Bucket.Name,
	}

	client_s3, _ := util.InitSession(&obj.Bucket)

	_, err := client_s3.CreateBucket(cparams)
	if err != nil {
		// Casting to the awserr.Error type will allow you to inspect the error
		// code returned by the service in code. The error code can be used
		// to switch on context specific functionality. In this case a context
		// specific error message is printed to the user based on the bucket
		// and key existing.
		//
		// For information on other S3 API error codes see:
		// http://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case aws_s3.ErrCodeBucketAlreadyExists:
				// pass
			case aws_s3.ErrCodeBucketAlreadyOwnedByYou:
				// pass
			default:
				log.Fatal(aerr)
			}
		}
	}
}

func (obj *S3) List(key string) ([]*aws_s3.Object, error) {

	// b, err := env.Env.GetBucket("extract")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	b := obj.Bucket
	//log.Println("Init S3 session for", b.Name, b.Credentials.Endpoint)
	// Always try and create the bucket.
	obj.CreateBucket()

	client_s3, _ := util.InitSession(&b)

	resp, err := client_s3.ListObjectsV2(&aws_s3.ListObjectsV2Input{Bucket: aws.String(b.Name)})
	if err != nil {
		log.Println(err)
		log.Fatal("COULD NOT LIST OBJECTS IN BUCKET ", b.Credentials.Endpoint, " ", b.Name)
	}

	keys := make([]*aws_s3.Object, 0)

	for _, item := range resp.Contents {
		// log.Println("Name:         ", *item.Key)
		// log.Println("Last modified:", *item.LastModified)
		// log.Println("Size:         ", *item.Size)
		// log.Println("Storage class:", *item.StorageClass)
		// log.Println("")

		// log.Printf("CHECKING OBJECT %s against filter %s\n", *item.Key, key)
		if found, _ := regexp.MatchString(key, *item.Key); found {
			keys = append(keys, item)
		}
	}
	return keys, nil
}

func (obj *S3) Get(key string) (map[string]string, error) {

	client, err := gokv_s3.NewClient(obj.Options)

	if err != nil {
		log.Println("s3 client")
		log.Fatal(err)
	}
	defer client.Close()

	json_map := make(map[string]string, 0)
	found, err := client.Get(key, &json_map)

	if err != nil {
		log.Println("s3 Get()")
		log.Fatal(err)
	}

	if found {
		return json_map, nil
	} else {
		return nil, fmt.Errorf("cannot get k/v for %s on bucket %s", key, obj.Bucket.Name)
	}
}

// https://github.com/nitisht/cookbook/blob/master/docs/aws-sdk-for-go-with-minio.md
func (obj *S3) PutObject(path []string, object []byte) {
	key := strings.Join(path, "/")
	b := obj.Bucket
	obj.CreateBucket()

	log.Printf("storing object at %s", key)
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#PutObjectInput
	_, err := obj.s3_client.PutObject(&s3.PutObjectInput{
		Body:        bytes.NewReader(object),
		Bucket:      &b.Name,
		Key:         aws.String(key),
		ContentType: aws.String(util.GetMimeType(path[len(path)-1])),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (obj *S3) GetObject(destination_filename string, key string) {
	b := obj.Bucket
	obj.CreateBucket()
	sess := obj.s3_session

	// 3) Create a new AWS S3 downloader
	downloader := s3manager.NewDownloader(sess)

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	file, err := os.Create(destination_filename)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(b.Name),
			Key:    aws.String(key),
		})

	if err != nil {
		log.Fatalf("Unable to download item %q, %v", key, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
}

func (obj *S3) StreamObject(filename string, path string) error {
	sess := obj.s3_session
	obj.CreateBucket()

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 2            // default is 5
	})

	//open the file
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("failed to open file %q, %v", filename, err)
		return err
	}
	//defer f.Close()

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(obj.Bucket.Name),
		Key:    aws.String(path),
		Body:   f,
	})

	//in case it fails to upload
	if err != nil {
		fmt.Printf("failed to upload file, %v", err)
		return err
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)
	return nil
}

func s3_client(b env.Bucket) (*s3.S3, *session.Session) {
	// https://stackoverflow.com/questions/41544554/how-to-run-aws-sdk-with-credentials-from-variables
	creds := credentials.NewStaticCredentials(
		b.Credentials.AccessKeyId,
		b.Credentials.SecretAccessKey,
		"")

	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String(b.Credentials.Endpoint),
		Region:      aws.String(b.Credentials.Region),
		Credentials: creds,
	})
	if err != nil {
		log.Fatal("CANNOT INIT AWS SESSION")
	}
	svc := s3.New(sess)
	return svc, sess
}

func minio_client(b env.Bucket) (*s3.S3, *session.Session) {
	// Configure to use MinIO Server

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			b.Credentials.AccessKeyId,
			b.Credentials.SecretAccessKey,
			""),
		Endpoint:         aws.String(b.Credentials.Endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess, err := session.NewSession(s3Config)
	if err != nil {
		log.Fatal("CANNOT INIT AWS SESSION")
	}

	s3Client := s3.New(sess)

	return s3Client, sess
}

// https://github.com/nitisht/cookbook/blob/master/docs/aws-sdk-for-go-with-minio.md

func initClient(b env.Bucket) (*s3.S3, *session.Session) {

	switch os.Getenv("ENV") {
	case "LOCAL":
		fallthrough
	case "DOCKER":
		client, session := minio_client(b)
		return client, session
	default:
		client, session := s3_client(b)
		return client, session
	}
}
