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
	Uri  string `mapstructure:"uri"`
	Port int    `mapstructure:"port"`
	// S3
	AccessKeyId     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	Endpoint        string `mapstructure:"endpoint"`
	// DB
	DbName   string `mapstructure:"db_name"`
	Host     string `mapstructure:"host"`
	Name     string `mapstructure:"name"`
	Password string `mapstructure:"password"`
	Username string `mapstructure:"username"`
}

type Service struct {
	Name        string      `mapstructure:"name"`
	Credentials Credentials `mapstructure:"credentials"`
}

type Database = Service
type Bucket = Service

type env struct {
	AppEnv       string               `mapstructure:"APPENV"`
	Home         string               `mapstructure:"HOME"`
	MemoryLimit  string               `mapstructure:"MEMORY_LIMIT"`
	Pwd          string               `mapstructure:"PWD"`
	TmpDir       string               `mapstructure:"TMPDIR"`
	User         string               `mapstructure:"USER"`
	VcapServices map[string][]Service `mapstructure:"VCAP_SERVICES"`

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
	viper.AddConfigPath("../../config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("can't find config files: ", err)
	}

	err = viper.Unmarshal(&Env)
	if err != nil {
		log.Fatal("environment can't be loaded: ", err)
	}

	// Configure the buckets and databases
	Env.Buckets = Env.VcapServices["s3"]
	Env.Databases = Env.VcapServices["aws-rds"]

}

func (e *env) GetServiceByName(category string, name string) (*Service, error) {
	for _, s := range e.VcapServices[category] {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("no service in category %s found with name %s", category, name)
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

func is_local_env() bool {
	return slices.Contains(local_envs, os.Getenv("ENV"))
}

func is_cf_env() bool {
	return slices.Contains(cf_envs, os.Getenv("ENV"))
}
