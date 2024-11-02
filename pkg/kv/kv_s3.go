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
	Bucket       env.Bucket
	MinioClient  *minio.Client
	MinioContext context.Context
}

type Storage interface {
	Store(string, JSON) error
	List(string) ([]*ObjInfo, error)
	Get(string) (Object, error)
}

// [{
// 	"resource": "/home/jadudm/git/search/eight/pkg/extract/process.go",
// 	"owner": "_generated_diagnostic_collection_name_#0",
// 	"code": {
// 		"value": "InvalidIfaceAssign",
// 		"target": {
// 			"$mid": 1,
// 			"path": "/golang.org/x/tools/internal/typesinternal",
// 			"scheme": "https",
// 			"authority": "pkg.go.dev",
// 			"fragment": "InvalidIfaceAssign"
// 		}
// 	},
// 	"severity": 8,
// 	"message": "cannot use s3_b (variable of type *procs.S3) as procs.Storage value in struct literal: *procs.S3 does not implement procs.Storage (wrong type for method Get)\n\t\thave Get(string) (procs.Object, error)\n\t\twant Get(string) (procs.JSON, error)",
// 	"source": "compiler",
// 	"startLineNumber": 23,
// 	"startColumn": 19,
// 	"endLineNumber": 23,
// 	"endColumn": 23
// }]

// Only open any given bucket once.
var buckets sync.Map

func NewKV(bucket_name string) *S3 {

	if s3, ok := buckets.Load(bucket_name); ok {
		return s3.(*S3)
	}

	s3 := S3{}

	// Grab a reference to our bucket from the config.
	b, err := env.Env.GetObjectStore(env.WorkingObjectStore)
	if err != nil {
		log.Fatal("ENV could not get bucket of name ", bucket_name)
	}
	s3.Bucket = b

	// Initialize minio client object.
	useSSL := true
	if env.IsContainerEnv() {
		useSSL = false
	}

	minioClient, err := minio.New(
		b.Credentials.Endpoint,
		&minio.Options{
			Creds: minio_credentials.NewStaticV4(
				b.Credentials.AccessKeyId,
				b.Credentials.SecretAccessKey, ""),
			Secure: useSSL,
		})
	if err != nil {
		log.Fatalln(err)
	}
	s3.MinioClient = minioClient
	s3.MinioContext = context.Background()

	found, err := minioClient.BucketExists(s3.MinioContext, bucket_name)
	if err != nil {
		log.Println("KV could not check if bucket exists ", bucket_name)
		log.Fatal(err)
	}

	if found {
		log.Println("KV found pre-existing bucket", bucket_name)
		return &s3
	}

	log.Println("KV creating new bucket ", bucket_name)
	// Try and make the bucket; if we're local, this is necessary.
	err = minioClient.MakeBucket(
		s3.MinioContext,
		bucket_name,
		minio.MakeBucketOptions{Region: b.Credentials.Region})

	if err != nil {
		log.Println(err)
		log.Fatal("KV could not create bucket ", bucket_name)
	}

	buckets.Store(bucket_name, &s3)

	return &s3
}

// GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions) (*Object, error)
func (s3 *S3) Get(key string) (Object, error) {
	object, err := s3.MinioClient.GetObject(
		s3.MinioContext,
		s3.Bucket.Name,
		key,
		minio.GetObjectOptions{})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return newJsonObjectFromMinio(key, object), nil
}

func (s3 *S3) GetFile(key string, dest_filename string) error {

	err := s3.MinioClient.FGetObject(
		context.Background(),
		s3.Bucket.Name,
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

	objectCh := s3.MinioClient.ListObjects(ctx, s3.Bucket.Name, minio.ListObjectsOptions{
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

	_, err := s3.MinioClient.PutObject(
		s3.MinioContext,
		s3.Bucket.Name,
		key,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: mime,
			PartSize:    5000000, // FIXME in bytes?
		},
	)
	if err != nil {
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
		log.Fatal(err)
	}

	return store(s3, key, fi.Size(), make(JSON, 0), reader)
}

// func (s3 *S3) Get(key string) (JSON, error) {

// 	client, err := gokv_s3.NewClient(obj.Options)

// 	if err != nil {
// 		log.Println("s3 client")
// 		log.Fatal(err)
// 	}
// 	defer client.Close()

// 	json_map := make(map[string]string, 0)
// 	found, err := client.Get(key, &json_map)

// 	if err != nil {
// 		log.Println("s3 Get()")
// 		log.Fatal(err)
// 	}

// 	if found {
// 		return json_map, nil
// 	} else {
// 		return nil, fmt.Errorf("cannot get k/v for %s on bucket %s", key, obj.Bucket.Name)
// 	}
// }

// https://github.com/nitisht/cookbook/blob/master/docs/aws-sdk-for-go-with-minio.md
// func (obj *S3) PutObject(path []string, object []byte) {
// 	key := strings.Join(path, "/")
// 	b := obj.Bucket
// 	obj.CreateBucket()

// 	log.Printf("storing object at %s", key)
// 	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#PutObjectInput
// 	_, err := obj.s3_client.PutObject(&s3.PutObjectInput{
// 		Body:        bytes.NewReader(object),
// 		Bucket:      &b.Name,
// 		Key:         aws.String(key),
// 		ContentType: aws.String(util.GetMimeType(path[len(path)-1])),
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func (obj *S3) GetObject(destination_filename string, key string) {
// 	b := obj.Bucket
// 	obj.CreateBucket()
// 	sess := obj.s3_session

// 	// 3) Create a new AWS S3 downloader
// 	downloader := s3manager.NewDownloader(sess)

// 	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
// 	file, err := os.Create(destination_filename)
// 	numBytes, err := downloader.Download(file,
// 		&s3.GetObjectInput{
// 			Bucket: aws.String(b.Name),
// 			Key:    aws.String(key),
// 		})

// 	if err != nil {
// 		log.Fatalf("Unable to download item %q, %v", key, err)
// 	}

// 	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
// }

// func (obj *S3) StreamObject(filename string, path string) error {
// 	sess := obj.s3_session
// 	obj.CreateBucket()

// 	// Create an uploader with the session and custom options
// 	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
// 		u.PartSize = 5 * 1024 * 1024 // The minimum/default allowed part size is 5MB
// 		u.Concurrency = 2            // default is 5
// 	})

// 	//open the file
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		fmt.Printf("failed to open file %q, %v", filename, err)
// 		return err
// 	}
// 	//defer f.Close()

// 	// Upload the file to S3.
// 	result, err := uploader.Upload(&s3manager.UploadInput{
// 		Bucket: aws.String(obj.Bucket.Name),
// 		Key:    aws.String(path),
// 		Body:   f,
// 	})

// 	//in case it fails to upload
// 	if err != nil {
// 		fmt.Printf("failed to upload file, %v", err)
// 		return err
// 	}
// 	fmt.Printf("file uploaded to, %s\n", result.Location)
// 	return nil
// }

////////////////////////////
// SUPPORT

func mapToReader(json_map JSON) (io.Reader, int64) {
	b, _ := json.Marshal(json_map)
	r := bytes.NewReader(b)
	return r, int64(len(b))
}
