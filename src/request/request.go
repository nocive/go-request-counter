// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package request

import (
	"time"
)

type Request struct {
	ClientIP  string    `json:"clientip"`
	Timestamp time.Time `json:"timestamp"`
}

func NewRequest(ip string) *Request {
	return &Request{
		ClientIP:  ip,
		Timestamp: time.Now(),
	}
}

func (this *Request) IsExpired(ttl time.Duration) bool {
	return time.Since(this.Timestamp).Nanoseconds() > ttl.Nanoseconds()
}
