// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package counter

import (
	"net"
	//"sync"
	//"sync/atomic"
	"time"
	"github.com/nocive/go-request-counter/src/request"
)

type RequestCounter struct {
	Requests map[string][]request.Request `json:"requests"`
	Ttl      time.Duration                `json:"ttl"`
	Count    int32                        `json:"count"`
}

func NewRequestCounter(t time.Duration) *RequestCounter {
	return &RequestCounter{
		Requests: make(map[string][]request.Request, 0),
		Ttl:      t,
		Count:    0,
	}
}

func (this *RequestCounter) Add(req request.Request) {
	this.Count++
	//atomic.AddInt32(&this.Count, 1)
	this.Requests[req.ClientIP] = append(this.Requests[req.ClientIP], req)
}

func (this *RequestCounter) Remove(ip string, idx int) {
	this.Requests[ip] = append(this.Requests[ip][:idx], this.Requests[ip][idx+1:]...)
	//atomic.AddInt32(&this.Count, -1)
	this.Count--
	if len(this.Requests[ip]) == 0 {
		delete(this.Requests, ip)
	}
}

func (this *RequestCounter) Mark(addr string) {
	ip, _, _ := net.SplitHostPort(addr)
	r := *request.NewRequest(ip)
	this.Add(r)
}

func (this *RequestCounter) Refresh() {
	// is this thread safe? or do we need a mutex?
	//var mux sync.Mutex
	//mux.Lock()
	for ip,reqs := range this.Requests {
		for i := len(reqs) - 1; i >= 0; i-- {
			if this.Requests[ip][i].IsExpired(this.Ttl) {
				this.Remove(ip, i)

			}
		}
	}
	//mux.Unlock()
}

func (this *RequestCounter) GetCount() int32 {
	return this.Count
}
