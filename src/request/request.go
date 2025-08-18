// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package request

import (
	"time"
)

type Request struct {
	ClientIP  string    `json:"-"`
	Timestamp time.Time `json:"timestamp"`
}

func NewRequest(ip string, ts time.Time) *Request {
	return &Request{
		ClientIP:  ip,
		Timestamp: ts,
	}
}

func (r *Request) IsExpired(ttl time.Duration) bool {
       return time.Since(r.Timestamp).Nanoseconds() > ttl.Nanoseconds()
}
