package pack

import (
	"context"
	"log"

	"search.eight/pkg/procs"
)

type Packer struct {
	Raw     map[string]string
	Storage procs.Storage
	Job     *PackRequestJob
}

type PackionFunction func(map[string]string)

func NewPacker(s procs.Storage, raw map[string]string, job *PackRequestJob) *Packer {
	return &Packer{
		Raw:     raw,
		Storage: s,
		Job:     job,
	}
}

func (e *Packer) Pack() {
	switch e.Raw["content-type"] {
	case "text/html":
		log.Println("HTML")
	case "application/pdf":
		log.Println("PDF")
	}
}

func (erw *PackRequestWorker) Work(
	ctx context.Context,
	job *PackRequestJob,
) error {
	log.Println("PACK", job.Args.Key)
	// Always safe to check the stats are ready.
	NewPackStats()

	// FIXME: Need a way to distinguish between  processing an entire
	// host domain and processing a single page?
	// objects, err := erw.FetchStorage.List(job.Args.Host)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("Found", len(objects), "objects")

	//for _, o := range objects { // use *o.Key as the path
	json_object, err := erw.FetchStorage.Get(job.Args.Key)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(json_object["path"], json_object["content-type"])
	e := NewPacker(erw.PackStorage, json_object, job)
	e.Pack()
	log.Println("PACK DONE", job.Args.Key)
	// }

	return nil
}
