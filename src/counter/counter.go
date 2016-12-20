// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package counter

import (
	"net"
	"time"

	"github.com/nocive/go-request-counter/src/request"
)

type RequestCounter struct {
	Ttl         time.Duration     `json:"ttl"`
	Requests    []request.Request `json:"requests"`
}

func NewRequestCounter(t time.Duration) *RequestCounter {
	return &RequestCounter{
		Ttl:      t,
		Requests: make([]request.Request, 0),
	}
}

func (this *RequestCounter) Add(req request.Request) {
	this.Requests = append(this.Requests, req)
}

func (this *RequestCounter) Remove(idx int) {
	this.Requests = append(this.Requests[:idx], this.Requests[idx+1:]...)
}

func (this *RequestCounter) Mark(ip net.IP) {
	this.Requests = append(this.Requests, *request.NewRequest(ip))
}

func (this *RequestCounter) Refresh() {
	for i := len(this.Requests) - 1; i >= 0; i-- {
		if this.Requests[i].Expired(this.Ttl) {
			this.Remove(i)
		}
	}
}

func (this *RequestCounter) Count() int {
	return len(this.Requests)
}
