package kv

import (
	"log"
	"os"
	"testing"

	"github.com/jadudm/eight/internal/env"
	"github.com/stretchr/testify/assert"
)

// Tests need a backend
// docker compose -f backend.yaml up

func setup( /* t *testing.T */ ) func(t *testing.T) {
	os.Setenv("ENV", "LOCALHOST")
	env.InitGlobalEnv()
	return func(t *testing.T) {
		// t.Log("teardown test case")
	}
}

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestKv(t *testing.T) {
	setup()
	log.Println(env.Env.ObjectStores)
	NewKV("fetch")
}

func TestName(t *testing.T) {
	setup()
	kv := NewKV("fetch")
	name := kv.Bucket.Name
	assert.Equal(t, "fetch", name)
}

func TestServices(t *testing.T) {
	setup()
	kv := NewKV("extract")
	assert.Equal(t, kv.Bucket.Name, "extract")
	s, err := env.Env.GetUserService("extract")
	if err != nil {
		t.Error(err)
	}
	log.Println(s)
	assert.Equal(t, "extract", s.Name)
}

func TestParams1(t *testing.T) {
	setup()
	kv := NewKV("extract")
	assert.Equal(t, kv.Bucket.Name, "extract")
	s, err := env.Env.GetUserService("extract")
	if err != nil {
		t.Error(err)
	}
	log.Println(s)
	assert.Equal(t, s.GetParamBool("extract_html"), true)
}

func TestParams2(t *testing.T) {
	setup()
	kv := NewKV("serve")
	assert.Equal(t, kv.Bucket.Name, "serve")
	s, err := env.Env.GetUserService("serve")
	if err != nil {
		t.Error(err)
	}
	log.Println(s)
	assert.Equal(t, "../../assets/databases", s.GetParamString("database_files_path"))
}

func TestParams3(t *testing.T) {
	setup()
	kv := NewKV("serve")
	assert.Equal(t, kv.Bucket.Name, "serve")
	s, err := env.Env.GetUserService("serve")
	if err != nil {
		t.Error(err)
	}
	log.Println(s)
	assert.Equal(t, int64(10004), s.GetParamInt64("external_port"))
}
