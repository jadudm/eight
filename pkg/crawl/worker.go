package crawl

import (
	"context"

	"github.com/riverqueue/river"
)

type CrawlRequestWorker struct {
	//VCAP *vcap.VcapServices
	// Out  chan *river.Job[CrawlRequest]
	Out chan *river.Job[CrawlRequest]

	// An embedded WorkerDefaults sets up default methods to fulfill the rest of
	// the Worker interface:
	river.WorkerDefaults[CrawlRequest]
}

// // https://github.com/philippgille/gokv?tab=readme-ov-file#usage

func (crw *CrawlRequestWorker) Work(
	ctx context.Context,
	job *river.Job[CrawlRequest],
) error {
	// log.Println("Running job", job.Args.Host, job.Args.Path, job.Queue)
	// crw.Config = s3.Options{
	// 	BucketName:         "gokv",
	// 	AWSaccessKeyID:     "foo",
	// 	AWSsecretAccessKey: "bar",
	// 	Region:             endpoints.UsWest2RegionID,
	// }
	//b := crw.VCAP.GetBucketByName("crawl-storage")
	// gokv_cfg := s3.Options{
	// 	BucketName:         b.ServiceName,
	// 	AWSaccessKeyID:     b.AccessKeyID,
	// 	AWSsecretAccessKey: b.SecretAccessKey,
	// 	Region:             b.Region,
	// }
	crw.Out <- job
	return nil
}
