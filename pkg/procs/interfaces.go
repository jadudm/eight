package procs

import aws_s3 "github.com/aws/aws-sdk-go/service/s3"

type Storage interface {
	Store(string, map[string]string) error
	Get(string) (map[string]string, error)
	List(string) ([]*aws_s3.Object, error)
}
