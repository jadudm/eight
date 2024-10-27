package procs

import (
	"log"

	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/s3"
	"search.eight/internal/env"
)

type S3 struct {
	Bucket  env.Bucket
	Options s3.Options
}

func NewKVS3(b env.Bucket) S3 {
	options := s3.Options{
		BucketName:             b.Name,
		Region:                 b.Credentials["region"].(string),
		CustomEndpoint:         b.Credentials["endpoint"].(string),
		UsePathStyleAddressing: true,
		AWSaccessKeyID:         b.Credentials["access_key_id"].(string),
		AWSsecretAccessKey:     b.Credentials["secret_access_key"].(string),
		Codec:                  encoding.JSON,
	}
	s3 := S3{
		Bucket:  b,
		Options: options,
	}

	log.Println("s3 init")
	return s3
}

func (obj S3) Store(key string, value any) error {

	client, err := s3.NewClient(obj.Options)

	if err != nil {
		log.Println("s3 client")
		log.Fatal(err)
	}

	key = key + ".json"
	err = client.Set(key, value)

	if err != nil {
		log.Println("s3 Set()")
		log.Fatal(err)
	}

	return nil
}
