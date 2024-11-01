package env

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/spf13/viper"
)

var Env *env

type env struct {
	AppEnv        string               `mapstructure:"APPENV"`
	Home          string               `mapstructure:"HOME"`
	MemoryLimit   string               `mapstructure:"MEMORY_LIMIT"`
	Pwd           string               `mapstructure:"PWD"`
	TmpDir        string               `mapstructure:"TMPDIR"`
	User          string               `mapstructure:"USER"`
	EightServices map[string][]Service `mapstructure:"EIGHT_SERVICES"`

	VcapServices map[string][]Service
	UserServices []Service
	Buckets      []Bucket
	Databases    []Database
}

type container_env struct {
	VcapServices map[string][]Service `mapstructure:"VCAP_SERVICES"`
}

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

type Parameters map[string]interface{}

type Service struct {
	Name        string      `mapstructure:"name"`
	Credentials Credentials `mapstructure:"credentials"`
	Parameters  Parameters  `mapstructure:"parameters"`
}

type Database = Service

type Bucket = Service

var container_envs = []string{"DOCKER", "GH_ACTIONS"}
var cf_envs = []string{"SANDBOX", "PREVIEW", "DEV", "STAGING", "PROD"}

func InitGlobalEnv() {
	Env = &env{}
	viper.AddConfigPath("/home/vcap/app/config")
	viper.SetConfigType("yaml")

	if IsContainerEnv() {
		log.Println("IsContainerEnv")
		viper.SetConfigName("container")
	}

	if IsCloudEnv() {
		log.Println("IsCloudEnv")
		viper.SetConfigName("cf")
		// https://github.com/spf13/viper/issues/1706
		// https://github.com/spf13/viper/issues/1671
		viper.AutomaticEnv()
	}

	err := viper.ReadInConfig()

	if err != nil {
		log.Fatal("ENV cannot load in the config file")
	}

	err = viper.Unmarshal(&Env)

	if err != nil {
		log.Fatal("ENV can't find config files: ", err)
	}

	if err != nil {
		log.Fatal("ENV environment can't be loaded: ", err)
	}

	// CF puts VCAP_* in a string containing JSON.
	// This means we don't have 1:1 locally *yet*, but
	// if we unpack things right, we end up with one struct
	// with everything in the rgiht places.
	if IsContainerEnv() {
		ContainerEnv := container_env{}
		viper.Unmarshal(&ContainerEnv)
		Env.VcapServices = ContainerEnv.VcapServices
	}

	if IsCloudEnv() {
		// CfEnv := cf_env{}
		// viper.Unmarshal(&CfEnv)
		new_vcs := make(map[string][]Service, 0)
		json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &new_vcs)
		Env.VcapServices = new_vcs
	}

	// Configure the buckets and databases
	Env.Buckets = Env.VcapServices["s3"]
	Env.Databases = Env.VcapServices["aws-rds"]
	Env.UserServices = Env.EightServices["user-provided"]

	for _, b := range Env.Buckets {
		log.Println(b.AsJson())
	}
}

// FIXME: I later added `GetService`, and it is a cleaner
// approach. Use that instead.
func (e *env) GetServiceByName(category string, name string) (*Service, error) {
	for _, s := range e.VcapServices[category] {
		if s.Name == name {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("ENV no service in category %s found with name %s", category, name)
}

// https://stackoverflow.com/questions/3582552/what-is-the-format-for-the-postgresql-connection-string-url
// postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]
func (e *env) GetDatabaseUrl(name string) (string, error) {
	params := ""
	if IsContainerEnv() {
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
	return "", fmt.Errorf("ENV no db found with name %s", name)
}

func (e *env) GetBucket(name string) (Bucket, error) {
	for _, b := range e.Buckets {
		if b.Name == name {
			return b, nil
		}
	}
	return Bucket{}, fmt.Errorf("ENV no bucket with name %s", name)
}

func (e *env) GetUserService(name string) (Service, error) {
	for _, s := range e.UserServices {
		if s.Name == name {
			return s, nil
		}
	}
	return Service{}, fmt.Errorf("ENV no service with name %s", name)
}

func IsContainerEnv() bool {
	return slices.Contains(container_envs, os.Getenv("ENV"))
}

func IsCloudEnv() bool {
	return slices.Contains(cf_envs, os.Getenv("ENV"))
}

func (s *Service) GetParamInt64(key string) int64 {
	if param_val, ok := s.Parameters[key]; ok {
		return int64(param_val.(int))
	} else {
		log.Fatalf("ENV no int param found for %s", key)
		return 0
	}
}

func (s *Service) GetParamString(key string) string {
	if param_val, ok := s.Parameters[key]; ok {
		return param_val.(string)
	} else {
		log.Fatalf("ENV no string param found for %s", key)
		return ""
	}
}

func (s *Service) GetParamBool(key string) bool {
	if param_val, ok := s.Parameters[key]; ok {
		return param_val.(bool)
	} else {
		log.Fatalf("ENV no bool param found for %s", key)
		return false
	}
}

func (s *Service) AsJson() string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	return string(b)
}
