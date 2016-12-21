// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package request

import (
	"sync"
	"sync/atomic"
	"time"
)

type RequestBucket struct {
	Requests map[string][]Request `json:"requests"`
	Ttl      time.Duration        `json:"ttl"`
	Count    int32                `json:"count"`

	//mux      sync.Mutex
}

func NewRequestBucket(t time.Duration) *RequestBucket {
	return &RequestBucket{
		Requests: make(map[string][]Request, 0),
		Ttl:      t,
		Count:    0,
	}
}

func (b *RequestBucket) Add(ip string, ts time.Time) {
	//var mux sync.Mutex
	//mux.Lock()
	//defer mux.Unlock()

	r := *NewRequest(ip, ts)
	b.Requests[ip] = append(b.Requests[ip], r)
	atomic.AddInt32(&b.Count, 1)
}

func (b *RequestBucket) AddNow(ip string) {
	b.Add(ip, time.Now())
}

func (b *RequestBucket) Remove(ip string, i int) {
	//var mux sync.Mutex
	//mux.Lock()
	//defer mux.Unlock()

	b.Requests[ip] = append(b.Requests[ip][:i], b.Requests[ip][i+1:]...)
	atomic.AddInt32(&b.Count, -1)

	if len(b.Requests[ip]) == 0 {
		delete(b.Requests, ip)
	}
}

func (b *RequestBucket) GetCount() int32 {
	return b.Count
}

func (b *RequestBucket) Refresh() int {
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()

	c := 0
	for ip,reqs := range b.Requests {
		for i := len(reqs) - 1; i >= 0; i-- {
			if b.Requests[ip][i].IsExpired(b.Ttl) {
				b.Remove(ip, i)
				c++
			}
		}
	}
	return c
}
