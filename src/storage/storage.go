// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nocive/go-request-counter/src/counter"
)

type RequestCounterStorage struct {
	path string
}

func NewRequestCounterStorage(p string) *RequestCounterStorage {
	return &RequestCounterStorage{
		path: filepath.Clean(p),
	}
}

func (this *RequestCounterStorage) Exists() bool {
	_, err := os.Stat(this.path)
	return !os.IsNotExist(err)
}

func (this *RequestCounterStorage) Create() error {
	file, err := os.Create(this.path)
	file.Close()
	return err
}

func (this *RequestCounterStorage) Save(c *counter.RequestCounter) error {
	file, err := os.Create(this.path)
	defer file.Close()
	if err != nil {
		return err
	}

	json, err := json.Marshal(c);
	if err == nil {
		file.WriteString(fmt.Sprintf("%s\n", json))
	}
	return err
}

func (this *RequestCounterStorage) Load(c *counter.RequestCounter) error {
	data, err := ioutil.ReadFile(this.path)
	if err == nil {
		err = json.Unmarshal(data, &c)
	}
	return err
}
