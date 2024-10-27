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
	return s3
}

func (obj S3) Store(key string, value map[string]string) error {

	client, err := s3.NewClient(obj.Options)

	if err != nil {
		log.Println("s3 client")
		log.Fatal(err)
	}
	defer client.Close()

	if obj.Options.Codec == encoding.JSON {
		key = key + ".json"
	} else if obj.Options.Codec == encoding.Gob {
		key = key + ".gob"
	}

	err = client.Set(key, value)

	if err != nil {
		log.Println("s3 Set()")
		log.Fatal(err)
	}

	return nil
}
