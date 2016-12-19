// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type RequestCounterStorage struct {
	path string
}

func NewRequestCounterStorage(p string) *RequestCounterStorage {
	return &RequestCounterStorage{
		path: p,
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

func (this *RequestCounterStorage) Save(c *RequestCounter) error {
	file, err := os.Create(this.path)
	defer file.Close()
	if err != nil {
		return err
	}

	if json, err := json.Marshal(c); err == nil {
		file.WriteString(fmt.Sprintf("%s\n", json))
	}
	return err
}

func (this *RequestCounterStorage) Load(c *RequestCounter) error {
	data, err := ioutil.ReadFile(this.path)
	if err == nil {
		err = json.Unmarshal(data, &c)
	}
	return err
}
