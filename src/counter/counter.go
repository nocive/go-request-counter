// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package counter

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"

	"github.com/nocive/go-request-counter/src/storage"
	"github.com/nocive/go-request-counter/src/request"
)

type RequestCounter struct {
	config  Config
	bucket  request.RequestBucket
	storage storage.RequestCounterStorage
	logger  logging.Logger
}

func NewRequestCounter(c Config, b request.RequestBucket, s storage.RequestCounterStorage, l logging.Logger) *RequestCounter {
	return &RequestCounter{
		config:  c,
		bucket:  b,
		storage: s,
		logger:  l,
	}
}

func (r *RequestCounter) Init() error {
	var err error

	r.logger.Infof("booting!")

	if !r.storage.Exists() {
		r.logger.Info("data file doesn't exist, creating")
		if err = r.storage.Create(); err != nil {
			return err
		}
	} else {
		r.logger.Info("data file exists, loading")
		if err = r.storage.Load(&r.bucket); err != nil {
			return err
		}

		c := r.bucket.GetCount()
		r.logger.Infof("%d requests loaded from data file", c)

		if (c > 0) {
			r.logger.Info("firing a manual refresh")
			c := r.bucket.Refresh() // purge expired events before starting
			r.logger.Infof("purged %d requests from bucket", c)
		}
	}

	r.logger.Infof(
		"[ttl %s] [refresh %s] [save %s] [sleep %s] [count %d]",
		r.config.RequestTtl,
		r.config.RefreshInterval,
		r.config.SaveInterval,
		r.config.SleepPerRequest,
		r.bucket.GetCount(),
	)

	r.logger.Info("initializing tickers and signal traps")
	r.traps()

	return nil
}

func (r *RequestCounter) Start() {
	sema := make(chan struct{}, r.config.MaxClients)

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()
		r.process(response, request)
	})

	r.logger.Infof("preparing to listen on %s", r.config.BindAddr)
	http.ListenAndServe(r.config.BindAddr, nil)
}

func (r *RequestCounter) process(response http.ResponseWriter, request *http.Request) {
	clientIP := request.Header.Get("X-Client-IP")
	if clientIP == "" {
		clientIP = request.RemoteAddr
		clientIP, _, _ = net.SplitHostPort(clientIP)
	}
	r.logger.Infof("request received [ip %s]", clientIP)

	r.bucket.AddNow(clientIP)
	currentCount := r.bucket.GetCount()

	io.WriteString(response, fmt.Sprintf("%d\n", currentCount))
	r.logger.Infof("counter incremented [count %d]", currentCount)

	time.Sleep(r.config.SleepPerRequest)
	r.logger.Info("request finished")
}

func (r *RequestCounter) shutdown() error {
	r.logger.Info("saving data to file")
	err := r.storage.Save(&r.bucket)
	return err
}

func (r *RequestCounter) traps() {
	quit := make(chan struct{})

	refreshTicker := time.NewTicker(r.config.RefreshInterval)
	saveTicker := time.NewTicker(r.config.SaveInterval)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		r.logger.Info("caught signal, cleaning up...")
		refreshTicker.Stop()
		saveTicker.Stop()
		if err := r.shutdown(); err != nil {
			r.logger.Fatal(err)
		}
		os.Exit(0)
	}()

	go func() {
		for {
			select {
			case <- refreshTicker.C:
				if c := r.bucket.GetCount(); c > 0 {
					r.logger.Infof("refresh [count %d]", c)
					r.bucket.Refresh()
				} else {
					r.logger.Info("-- idling --")
				}

			case <- saveTicker.C:
				r.logger.Info("snapshotting data to file")
				r.storage.Save(&r.bucket)

			case <- quit:
				refreshTicker.Stop()
				saveTicker.Stop()
				return
			}
		}
	}()
}

