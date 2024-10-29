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

	cfg := &aws.Config{
		Credentials: creds,
		Endpoint:    aws.String(b.Credentials.Endpoint),
		Region:      aws.String(b.Credentials.Region),
	}

	// Set additional properties when running in a containerized
	// environment... SSL is disabled in local/GH containers.
	if env.IsContainerEnv() {
		cfg.DisableSSL = aws.Bool(true)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal("CANNOT INIT AWS SESSION")
	}
	svc := s3.New(sess)
	return svc, sess
}
