package env

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/spf13/viper"
)

type Credentials struct {
	// Common
	Uri  string
	Port int
	// S3
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	Bucket          string
	Endpoint        string
	// DB
	DbName   string
	Host     string
	Name     string
	Password string
	Username string
}
type Bucket struct {
	Name        string
	Credentials Credentials
}

type Database struct {
	Name        string
	Credentials Credentials
}
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

var local_envs = []string{"DOCKER", "GH_ACTIONS"}
var cf_envs = []string{"PREVIEW", "DEV", "STAGING", "PROD"}

func InitGlobalEnv() {
	Env = &env{}

	if is_local_env() {
		viper.SetConfigName("local")
	}
	if is_cf_env() {
		viper.SetConfigName("cfenv")
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath("/home/vcap/app/config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("can't find the file .env : ", err)
	}

	err = viper.Unmarshal(&Env)
	if err != nil {
		log.Fatal("environment can't be loaded: ", err)
	}

	// Configure the buckets and databases
	configureBuckets()
	configureDatabases()

}

// https://stackoverflow.com/questions/3582552/what-is-the-format-for-the-postgresql-connection-string-url
// postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]
func (e *env) GetDatabaseUrl(name string) (string, error) {
	params := ""
	if is_local_env() {
		params = "sslmode=disable"
	}
	for _, db := range e.Databases {
		if db.Name == name {
			return fmt.Sprintf("postgresql://%s@%s:%d/%s?%s",
				db.Credentials.Username,
				db.Credentials.Host,
				db.Credentials.Port,
				db.Credentials.DbName,
				params,
			), nil
		}
	}
	return "", fmt.Errorf("no db found with name %s", name)
}

func (e *env) GetBucket(name string) (Bucket, error) {
	for _, b := range e.Buckets {
		if b.Name == name {
			return b, nil
		}
	}
	return Bucket{}, fmt.Errorf("no bucket with name %s", name)
}

func configureBuckets() {
	vcap := viper.Get("VCAP_SERVICES").(map[string]interface{})
	for _, b := range vcap["s3"].([]interface{}) {
		b := b.(map[string]interface{})
		creds := b["credentials"].(map[string]interface{})
		bucket := Bucket{
			Name: b["name"].(string),
			Credentials: Credentials{
				Uri:             creds["uri"].(string),
				Port:            creds["port"].(int),
				AccessKeyId:     creds["access_key_id"].(string),
				SecretAccessKey: creds["secret_access_key"].(string),
				Region:          creds["region"].(string),
				Bucket:          creds["bucket"].(string),
				Endpoint:        creds["endpoint"].(string),
			},
		}
		Env.Buckets = append(Env.Buckets, bucket)
	}
}

func is_local_env() bool {
	return slices.Contains(local_envs, os.Getenv("ENV"))
}

func is_cf_env() bool {
	return slices.Contains(cf_envs, os.Getenv("ENV"))
}

func configureDatabases() {
	vcap := viper.Get("VCAP_SERVICES").(map[string]interface{})
	for _, b := range vcap["aws-rds"].([]interface{}) {
		b := b.(map[string]interface{})
		creds := b["credentials"].(map[string]interface{})
		db := Database{
			Name: b["name"].(string),
			Credentials: Credentials{
				DbName:   creds["db_name"].(string),
				Port:     creds["port"].(int),
				Name:     creds["name"].(string),
				Password: creds["password"].(string),
				Username: creds["username"].(string),
				Uri:      creds["uri"].(string),
				Host:     creds["host"].(string),
			},
		}
		Env.Databases = append(Env.Databases, db)
	}
}
