// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nocive/go-request-counter/src/request"
)

type RequestCounterStorage struct {
	path string
}

func NewRequestCounterStorage(p string) *RequestCounterStorage {
	return &RequestCounterStorage{
		path: filepath.Clean(p),
	}
}

func (s *RequestCounterStorage) Exists() bool {
	_, err := os.Stat(s.path)
	return !os.IsNotExist(err)
}

func (s *RequestCounterStorage) Create() error {
	file, err := os.Create(s.path)
	file.Close()
	return err
}

func (s *RequestCounterStorage) Save(b *request.RequestBucket) error {
	file, err := os.Create(s.path)
	defer file.Close()
	if err != nil {
		return err
	}

	json, err := json.Marshal(b);
	if err == nil {
		file.WriteString(fmt.Sprintf("%s\n", json))
	}
	return err
}

func (s *RequestCounterStorage) Load(b *request.RequestBucket) error {
	data, err := ioutil.ReadFile(s.path)
	if err == nil {
		err = json.Unmarshal(data, &b)
	}
	return err
}
