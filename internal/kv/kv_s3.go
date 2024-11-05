package kv

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
	"go.uber.org/zap"

	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/util"
)

// Only open any given bucket once.
var buckets sync.Map

type S3 struct {
	Bucket      env.Bucket
	MinioClient *minio.Client
}

type Storage interface {
	Store(string, JSON) error
	List(string) ([]*ObjInfo, error)
	Get(string) (Object, error)
}

func NewKV(bucket_name string) S3 {
	if !env.IsValidBucketName(bucket_name) {
		log.Fatal("KV INVALID BUCKET NAME ", bucket_name)
	}

	// Check if we already have this in the map, so reconnects don't create
	// new S3 objects/etc.
	if s3, ok := buckets.Load(bucket_name); ok {
		zap.L().Debug("in the sync map", zap.String("bucket_name", bucket_name))
		return s3.(S3)
	} else {
		zap.L().Debug("not in the sync map", zap.String("bucket_name", bucket_name))
	}

	s3 := S3{}

	// Grab a reference to our bucket from the config.
	b, err := env.Env.GetObjectStore(bucket_name)

	if err != nil {
		zap.L().Error("could not get bucket from config", zap.String("bucket_name", bucket_name))
		os.Exit(1)
	}

	zap.L().Debug("got reference to bucket from vcap",
		zap.String("name", b.Name),
		zap.String("bucket", b.CredentialString("bucket")),
		zap.String("region", b.CredentialString("region")))

	s3.Bucket = b

	// Initialize minio client object.
	useSSL := true
	if env.IsContainerEnv() || env.IsLocalTestEnv() {
		// log.Println("ENV disabling SSL in containerized environment")
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
		//log.Println("KV could not check if bucket exists ", bucket_name)
		//log.Fatal(err)
		zap.L().Fatal("could not check if bucket exists", zap.String("bucket_name", bucket_name))
	}

	if found {
		zap.L().Debug("pre-existing bucket in S3",
			zap.String("bucket_name", bucket_name))
		// Make sure to insert the metadata into the sync.Map
		// when we find a bucket that already exists.
		buckets.Store(bucket_name, s3)
		zap.L().Info("found pre-existing bucket in S3")
		return s3
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
	} // Skip container creation in CF

	// Put a pointer to this object in our syncmap.
	buckets.Store(bucket_name, &s3)

	loaded, _ := buckets.Load(bucket_name)
	zap.L().Info("bucket ready", zap.String("bucket_name", loaded.(*S3).Bucket.Name))

	return s3
}

// GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions) (*Object, error)
func (s3 *S3) Get(key string) (Object, error) {
	ctx := context.Background()
	bucket_name := s3.Bucket.CredentialString("bucket")

	// The object has a channel interface that we have to empty.
	object, err := s3.MinioClient.GetObject(
		ctx,
		bucket_name,
		key,
		minio.GetObjectOptions{})

	zap.L().Debug("retrieved S3 object", zap.String("key", key))

	if err != nil {
		log.Println(s3.Bucket.CredentialString("bucket"), key)
		log.Println(err)
		return nil, err
	}

	raw, err := io.ReadAll(object)
	if err != nil {
		log.Fatal("KV could not read object bytes ", bucket_name, " ", key)
	}
	jsonm := make(JSON)
	json.Unmarshal(raw, &jsonm)
	mime := "octet/binary"
	if v, ok := jsonm["content-type"]; ok {
		mime = util.CleanMimeType(v)
	}
	return Obj{
		info: &ObjInfo{
			Key:  key,
			Size: int64(len(raw)),
			Mime: mime,
		},
		value: jsonm,
	}, nil
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

	objectCh := s3.MinioClient.ListObjects(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		minio.ListObjectsOptions{
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
func store(s3 *S3, destination_key string, size int64, jsonm JSON, reader io.Reader) error {
	mime := "octet/binary"

	if jsonm != nil {
		mime = "application/json"
	}

	ctx := context.Background()
	log.Println("KV store", s3.Bucket.Name, destination_key, size)
	_, err := s3.MinioClient.PutObject(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		destination_key,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: mime,
			// This seems to set the *minimum* partsize for multipart uploads.
			// Which... makes writing JSON objects impossible.
			// PartSize:    5000000
		},
	)
	if err != nil {
		log.Println("KV cannot store", destination_key, size, jsonm)
		log.Println(err)
	}
	return err

}

func (s3 *S3) Store(key string, jsonm JSON) error {
	reader, size := mapToReader(jsonm)
	return store(s3, key, size, jsonm, reader)
}

func (s3 *S3) StoreFile(destination_key string, source_filename string) error {
	reader, err := os.Open(source_filename)
	if err != nil {
		log.Fatal("KV cannot open file", source_filename)
	}
	fi, err := reader.Stat()
	if err != nil {
		log.Println("KV could not stat file")
		log.Fatal(err)
	}

	return store(s3, destination_key, fi.Size(), make(JSON, 0), reader)
}

////////////////////////////
// SUPPORT

func mapToReader(json_map JSON) (io.Reader, int64) {
	b, _ := json.Marshal(json_map)
	r := bytes.NewReader(b)
	return r, int64(len(b))
}
