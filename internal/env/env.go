package env

import (
	"errors"
	"fmt"
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/spf13/viper"
)

type Bucket = cfenv.Service
type Database = cfenv.Service

type env struct {
	AppEnv      string `mapstructure:"APPENV"`
	Home        string `mapstructure:"HOME"`
	MemoryLimit string `mapstructure:"MEMORY_LIMIT"`
	Pwd         string `mapstructure:"PWD"`
	TmpDir      string `mapstructure:"TMPDIR"`
	User        string `mapstructure:"USER"`

	Buckets   []Bucket
	Databases []Database
}

var Env *env

func InitGlobalEnv() {
	Env = &env{}

	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("can't find the file .env : ", err)
	}

	err = viper.Unmarshal(&Env)
	if err != nil {
		log.Fatal("environment can't be loaded: ", err)
	}
}

func (e *env) GetBucket(name string) (Bucket, error) {
	for _, b := range e.Buckets {
		if b.Name == name {
			return b, nil
		}
	}
	return Bucket{}, errors.New(fmt.Sprintf("no bucket with name %s", name))
}

// func NewFromFile(vcap_json string) *Env {
// 	env := Env{}

// 	// VCAP_APPLICATION
// 	os.Setenv("VCAP_APPLICATION", "{}")
// 	// Required environment variables
// 	os.Setenv("HOME", "/home/vcap/app")
// 	os.Setenv("MEMORY_LIMIT", "512m")
// 	os.Setenv("PWD", "/home/vcap")
// 	os.Setenv("TMPDIR", "/home/vcap/tmp")
// 	os.Setenv("USER", "vcap")

// 	// VCAP_SERVICES
// 	js, _ := os.ReadFile(vcap_json)
// 	var vcap_services map[string]interface{}
// 	err := json.Unmarshal(js, &vcap_services)

// 	if err != nil {
// 		log.Println("VCAP_SERVICES")
// 		log.Fatal(err)
// 	}

// 	b, _ := json.Marshal(vcap_services)
// 	os.Setenv("VCAP_SERVICES", string(b))

// 	app, err := cfenv.Current()

// 	if err != nil {
// 		log.Println("cfenv.Current")
// 		log.Fatal(err)
// 	}

// 	env.Buckets = app.Services["s3"]
// 	env.Databases = app.Services["aws-rds"]

// 	log.Println("Buckets: ", len(env.Buckets))
// 	log.Println("Databases: ", len(env.Databases))
// 	return &env
// }

/*
	app.Service

	[
	  {
	    "Name": "backups",
	    "Label": "s3",
	    "Tags": [
	      "AWS",
	      "S3",
	      "object-storage"
	    ],
	    "Plan": "basic",
	    "Credentials": {
	      "access_key_id": "nutnutnut",
	      "additional_buckets": [],
	      "bucket": "ephemeral-storage",
	      "endpoint": "minio",
	      "fips_endpoint": "minio",
	      "insecure_skip_verify": false,
	      "port": 9000,
	      "region": "us-east-1",
	      "secret_access_key": "nutnutnut",
	      "uri": "http://minio:9000"
	    },
	    "VolumeMounts": null
	  }
	]
*/
