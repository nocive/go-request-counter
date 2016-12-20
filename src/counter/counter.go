// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package counter

import (
	"time"
)

type RequestCounter struct {
	Ttl      time.Duration `json:"ttl"`
	Requests []int64       `json:"requests"`
}

func NewRequestCounter(t time.Duration) *RequestCounter {
	return &RequestCounter{
		Ttl:      t,
		Requests: make([]int64, 0),
	}
}

func (this *RequestCounter) Add(ts int64) {
	this.Requests = append(this.Requests, ts)
}

func (this *RequestCounter) Remove(idx int) {
	this.Requests = append(this.Requests[:idx], this.Requests[idx+1:]...)
}

func (this *RequestCounter) Mark() {
	this.Requests = append(this.Requests, time.Now().UnixNano())
}

func (this *RequestCounter) Expired(ts int64) bool {
	t := time.Unix(0, ts)
	elapsed := time.Since(t)

	return elapsed.Seconds() > this.Ttl.Seconds()
}

func (this *RequestCounter) Refresh() {
	for i := len(this.Requests) - 1; i >= 0; i-- {
		if this.Expired(this.Requests[i]) {
			this.Remove(i)
		}
	}
}

func (this *RequestCounter) Count() int {
	return len(this.Requests)
}
