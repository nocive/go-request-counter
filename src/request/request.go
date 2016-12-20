// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package request

import (
	"net"
	"time"
)

type Request struct {
	ClientIP  net.IP    `json:"clientip"`
	Timestamp time.Time `json:"timestamp"`
}

func NewRequest(ip net.IP) *Request {
	return &Request{
		ClientIP:  ip,
		Timestamp: time.Now(),
	}
}

func (this *Request) Expired(ttl time.Duration) bool {
	return time.Since(this.Timestamp).Nanoseconds() > ttl.Nanoseconds()
}
