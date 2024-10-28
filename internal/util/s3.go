package util

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"search.eight/internal/env"
)

// func session_s3(b *env.Bucket) (*aws_s3.S3, *session.Session) {
func InitSession(b *env.Bucket) (*s3.S3, *session.Session) {

	// https://stackoverflow.com/questions/41544554/how-to-run-aws-sdk-with-credentials-from-variables
	creds := credentials.NewStaticCredentials(b.Credentials.AccessKeyId, b.Credentials.SecretAccessKey, "")

	var cfg *aws.Config

	if env.IsLocalEnv() {
		cfg = &aws.Config{
			Credentials:      creds,
			Endpoint:         aws.String(b.Credentials.Endpoint),
			Region:           aws.String(b.Credentials.Region),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		}
	} else if env.IsCloudEnv() {
		cfg = &aws.Config{
			Credentials: creds,
			Endpoint:    aws.String(b.Credentials.Endpoint),
			Region:      aws.String(b.Credentials.Region),
		}
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal("CANNOT INIT AWS SESSION")
	}
	svc := s3.New(sess)
	return svc, sess
}
