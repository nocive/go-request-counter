// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package main

import (
	"encoding/json"
	"time"

	"github.com/likexian/simplejson-go"
)

type RequestCounter struct {
	ttl        time.Duration
	timestamps []int64
}

func NewRequestCounter(t time.Duration) *RequestCounter {
	return &RequestCounter{
		ttl:        t,
		timestamps: make([]int64, 0),
	}
}

func (this *RequestCounter) Add(ts int64) {
	this.timestamps = append(this.timestamps, ts)
}

func (this *RequestCounter) Remove(idx int) {
	this.timestamps = append(this.timestamps[:idx], this.timestamps[idx+1:]...)
}

func (this *RequestCounter) Mark() {
	this.timestamps = append(this.timestamps, time.Now().UnixNano())
}

func (this *RequestCounter) Expired(ts int64) bool {
	t := time.Unix(0, ts)
	elapsed := time.Since(t)

	return elapsed.Seconds() > this.ttl.Seconds()
}

func (this *RequestCounter) Refresh() {
	for i := len(this.timestamps) - 1; i >= 0; i-- {
		if this.Expired(this.timestamps[i]) {
			this.Remove(i)
		}
	}
}

func (this *RequestCounter) Count() int {
	return len(this.timestamps)
}

func (this *RequestCounter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
            "ttl":        this.ttl,
            "timestamps": this.timestamps,
        })
}

func (this *RequestCounter) UnmarshalJSON(data []byte) (err error) {
	json, err := simplejson.Loads(string(data))
	if err != nil {
		return err
	}

	timestamps, err := json.Get("timestamps").Array()
	if err != nil {
		return err
	}

	for _,ts := range timestamps {
		this.timestamps = append(this.timestamps, (int64)(ts.(float64)))
	}

	return nil
}
