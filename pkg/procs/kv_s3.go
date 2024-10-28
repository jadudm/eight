package procs

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	aws_s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/philippgille/gokv/encoding"
	gokv_s3 "github.com/philippgille/gokv/s3"

	"search.eight/internal/env"
	"search.eight/internal/util"
)

type S3 struct {
	Bucket  env.Bucket
	Options gokv_s3.Options
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
		return nil, fmt.Errorf("cannot get k/v for %s", key)
	}
}
