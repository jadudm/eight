package procs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	minio "github.com/minio/minio-go/v7"
	minio_credentials "github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/jadudm/eight/internal/env"
)

type S3 struct {
	Bucket      env.Bucket
	MinioClient *minio.Client
}

type Storage interface {
	Store(string, JSON) error
	List(string) ([]*ObjInfo, error)
	Get(string) (Object, error)
}

// Only open any given bucket once.
var buckets sync.Map

func NewKV(bucket_name string) *S3 {

	if s3, ok := buckets.Load(bucket_name); ok {
		return s3.(*S3)
	}

	s3 := S3{}

	// Grab a reference to our bucket from the config.
	b, err := env.Env.GetObjectStore(bucket_name)
	if err != nil {
		log.Fatal("ENV could not get bucket of name ", bucket_name)
	}
	s3.Bucket = b

	// Initialize minio client object.
	useSSL := true
	if env.IsContainerEnv() {
		log.Println("ENV disabling SSL in containerized environment")
		useSSL = false
	}

	options := minio.Options{
		Creds: minio_credentials.NewStaticV4(
			b.CredentialString("access_key_id"),
			b.CredentialString("secret_access_key"), ""),
		Secure: useSSL,
	}

	minioClient, err := minio.New(b.CredentialString("endpoint"), &options)
	if err != nil {
		log.Fatalln(err)
	}
	s3.MinioClient = minioClient
	ctx := context.Background()

	found, err := minioClient.BucketExists(ctx, s3.Bucket.CredentialString("bucket"))
	if err != nil {
		log.Println("KV could not check if bucket exists ", bucket_name)
		log.Fatal(err)
	}

	if found {
		log.Println("KV found pre-existing bucket", bucket_name)
		return &s3
	}

	if env.IsContainerEnv() {
		log.Println("KV creating new bucket ", bucket_name)
		// Try and make the bucket; if we're local, this is necessary.
		ctx := context.Background()
		err = minioClient.MakeBucket(
			ctx,
			s3.Bucket.CredentialString("bucket"),
			minio.MakeBucketOptions{Region: b.CredentialString("region")})

		if err != nil {
			log.Println(err)
			log.Fatal("KV could not create bucket ", bucket_name)
		}
	} else {
		log.Println("KV skipping bucket creation in cloud env")
	}

	buckets.Store(bucket_name, &s3)

	return &s3
}

// GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions) (*Object, error)
func (s3 *S3) Get(key string) (Object, error) {
	ctx := context.Background()
	object, err := s3.MinioClient.GetObject(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		key,
		minio.GetObjectOptions{})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return newJsonObjectFromMinio(key, object), nil
}

func (s3 *S3) GetFile(key string, dest_filename string) error {
	ctx := context.Background()
	err := s3.MinioClient.FGetObject(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		key,
		dest_filename,
		minio.GetObjectOptions{})

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// //////////////////
// LIST
// Lists objects in the bucket, returning keys and sizes.
func (s3 *S3) List(prefix string) ([]*ObjInfo, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := s3.MinioClient.ListObjects(ctx, s3.Bucket.CredentialString("bucket"), minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	})

	objects := make([]*ObjInfo, 0)
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return nil, object.Err
		}
		objects = append(objects, NewObjInfo(object.Key, object.Size))
	}
	return objects, nil
}

// //////////////////
// STORE
// Stores a k,v to the bucket
func store(s3 *S3, key string, size int64, jsonm JSON, reader io.Reader) error {
	mime := "octet/binary"
	if v, ok := jsonm["content-type"]; ok {
		mime = v
	}
	ctx := context.Background()
	log.Println("KV store", s3.Bucket.Name, key, size)
	_, err := s3.MinioClient.PutObject(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		key,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: mime,
			PartSize:    5000000, // FIXME in bytes?
		},
	)
	if err != nil {
		log.Println("KV cannot store", key, size, jsonm)
		log.Println(err)
	}
	return err

}

func (s3 *S3) Store(key string, jsonm JSON) error {
	reader, size := mapToReader(jsonm)
	return store(s3, key, size, jsonm, reader)
}

func (s3 *S3) StoreFile(key string, filename string) error {
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal("KV cannot open file", filename)
	}
	fi, err := reader.Stat()
	if err != nil {
		log.Println("KV could not stat file")
		log.Fatal(err)
	}

	return store(s3, key, fi.Size(), make(JSON, 0), reader)
}

////////////////////////////
// SUPPORT

func mapToReader(json_map JSON) (io.Reader, int64) {
	b, _ := json.Marshal(json_map)
	r := bytes.NewReader(b)
	return r, int64(len(b))
}
