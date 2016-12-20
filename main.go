// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package main

import (
	"io"
	"fmt"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nocive/go-request-counter/src/counter"
	"github.com/nocive/go-request-counter/src/storage"
)

func handleError(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Fatal(fmt.Sprintf("Line: %d\tFile: %s\nMessage: %s", line, file, err))
		os.Exit(1)
	}
}

func boot() {
	var err error

	log.Printf("boot :: booting! ttl: %d / refresh interval: %d / save interval: %d\n", ttl, refreshInterval, saveInterval)

	if !stg.Exists() {
		log.Println("boot :: data file doesn't exist, creating")
		err = stg.Create()
		handleError(err)
		log.Println("boot :: starting with empty counter")
	} else {
		log.Println("boot :: data file exists, loading")
		err = stg.Load(cnt)
		handleError(err)

		log.Printf("boot :: %d requests loaded from data file\n", cnt.Count())
		log.Println("boot :: firing a manual refresh")
		cnt.Refresh() // purge expired events before starting
	}

	log.Println("boot :: initializing signal traps")
	trap()

	log.Println("boot :: initializing ticker")
	ticker()
}

func start() {
	sema := make(chan struct{}, maxClients)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()

		log.Println("serve :: request received")

		cnt.Mark(net.ParseIP(r.RemoteAddr))
		currentCount := cnt.Count()

		io.WriteString(w, fmt.Sprintf("%d\n", currentCount))
		log.Printf("serve :: hits incremented! current: %d\n", currentCount)

		time.Sleep(2 * time.Second)

		log.Println("serve :: request finished")
	})

	log.Println("start :: preparing to listen on", bind)
	http.ListenAndServe(bind, nil)
}

func shutdown() {
	log.Println("shutdown :: saving data to file")

	err := stg.Save(cnt)
	handleError(err)
}


func trap() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("trap :: caught signal, cleaning up...")
		shutdown()
		os.Exit(0)
	}()
}

func ticker() {
	quit := make(chan struct{})

	refreshTicker := time.NewTicker(time.Duration(refreshInterval) * time.Millisecond)
	saveTicker := time.NewTicker(time.Duration(saveInterval) * time.Millisecond)

	go func() {
		for {
			select {
			case <- refreshTicker.C:
				log.Printf("ticker :: refresh: %d\n", cnt.Count())
				cnt.Refresh()

			case <- saveTicker.C:
				log.Println("ticker :: snapshoting data to file")
				stg.Save(cnt)

			case <- quit:
				refreshTicker.Stop()
				saveTicker.Stop()
				return
			}
		}
	}()
}

func main() {
	flag.StringVar(&bind, "bind", defaultBindAddr, "which address to bind to in the form of addr:port.")
	flag.IntVar(&ttl, "ttl", defaultRequestTtl, "request ttl in seconds. when requests expire they are no longer counted.")
	flag.StringVar(&path, "path", defaultDataPath, "path to the storage filename.")
	flag.BoolVar(&help, "help", false, "display this help text.")
	flag.Parse()

	if help {
		fmt.Printf("Usage: %s [-bind address] [-ttl ttl] [-path path]\n\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	if string(path[0]) != "/" {
		wd, err := os.Getwd()
		handleError(err)
		path = fmt.Sprintf("%s/%s", wd, path)
	}

	cnt = counter.NewRequestCounter(time.Duration(ttl) * time.Millisecond)
	stg = storage.NewRequestCounterStorage(path)

	boot()
	start()
}

const (
	// default path + filename where to store the data in
	defaultDataPath   = "data/counter.json"

	// default bind address to listen to
	defaultBindAddr   = "0.0.0.0:6666"

	// default request time to live
	defaultRequestTtl = 60000

	// max concurrent clients allowed
	maxClients        = 5

	// how often should the counter data be refreshed
	refreshInterval   = 1000

	// how often should the counter data be saved to disk
	saveInterval      = 90000
)

var (
	// cli arguments
	bind  string
	ttl   int
	path  string
	help  bool

	cnt   *counter.RequestCounter
	stg   *storage.RequestCounterStorage
)
