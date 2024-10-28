package extract

import (
	"context"
	"log"
)

func (crw *ExtractRequestWorker) Work(
	ctx context.Context,
	job *ExtractRequestJob,
) error {
	log.Println("EXTRACT", job)
	return nil
}
