package procs

import (
	"encoding/json"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
)

type JSON map[string]string

type Object interface {
	GetKey() string
	GetJson() JSON
	GetValue(string) string
	GetSize() int64
	GetMimeType() string
}

type ObjInfo struct {
	key  string
	size int64
	mime string
}

func NewObjInfo(key string, size int64) *ObjInfo {
	return &ObjInfo{
		key:  key,
		size: size,
	}
}

type Obj struct {
	value JSON
	info  *ObjInfo
	// bytes []byte
}

func newJsonObjectFromMinio(key string, mo *minio.Object) Obj {
	raw, err := io.ReadAll(mo)
	if err != nil {
		log.Fatal("KV could not read object bytes")
	}
	jsonm := make(JSON)
	json.Unmarshal(raw, &jsonm)
	mime := "octet/binary"
	if v, ok := jsonm["content-type"]; ok {
		mime = v
	}
	return Obj{
		info: &ObjInfo{
			key:  key,
			size: int64(len(raw)),
			mime: mime,
		},
		value: jsonm,
	}
}

func NewObject(key string, value JSON) *Obj {
	b, err := json.Marshal(value)
	if err != nil {
		log.Fatal("ENV could not marshal", key)
	}

	size := int64(len(b))
	mime := ""
	if good, ok := value["content-type"]; !ok {
		mime = "octet/binary"
	} else {
		mime = good
	}

	return &Obj{
		info: &ObjInfo{
			key:  key,
			size: size,
			mime: mime,
		},
		value: value,
	}
}

func (o Obj) GetKey() string {
	return o.info.key
}

func (o Obj) GetValue(key string) string {
	return o.value[key]
}

func (o Obj) GetJson() JSON {
	return o.value
}

func (o Obj) GetSize() int64 {
	return o.info.size
}

func (o Obj) GetMimeType() string {
	return o.info.mime
}
