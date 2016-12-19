// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package main

import (
	"io"
	"fmt"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
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

	log.Printf("boot :: booting! ttl: %d / refresh interval: %d / save interval: %d\n", ttl, rInterval, sInterval)

	if !storage.Exists() {
		log.Println("boot :: data file doesn't exist, creating")
		err = storage.Create()
		handleError(err)
		log.Println("boot :: starting with empty counter")
	} else {
		log.Println("boot :: data file exists, loading")
		err = storage.Load(counter)
		handleError(err)

		log.Printf("boot :: %d requests loaded from data file\n", counter.Count())
		log.Println("boot :: firing a manual refresh")
		counter.Refresh() // purge expired events before starting
	}

	log.Println("boot :: initializing signal traps")
	trap()

	log.Println("boot :: initializing ticker")
	ticker()
}

func start() {
	http.HandleFunc("/", serve)
	log.Println("start :: preparing to listen on", bind)
	http.ListenAndServe(bind, nil)
}

func shutdown() {
	log.Println("shutdown :: saving data to file")

	err := storage.Save(counter)
	handleError(err)
}

func serve(w http.ResponseWriter, r *http.Request) {
	log.Println("serve :: request received")

	counter.Mark()
	currentCount := counter.Count()

	io.WriteString(w, fmt.Sprintf("%d\n", currentCount))
	log.Printf("serve :: hits incremented! current: %d\n", currentCount)

	log.Println("serve :: request finished")
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

	rticker := time.NewTicker(time.Duration(rInterval) * time.Second)
	go func() {
		for {
			select {
			case <- rticker.C:
				log.Printf("ticker :: refresh: %d\n", counter.Count())
				counter.Refresh()
			case <- quit:
				rticker.Stop()
				return
			}
		}
	}()

	sticker := time.NewTicker(time.Duration(sInterval) * time.Second)
	go func() {
		for {
			select {
			case <- sticker.C:
				log.Println("ticker :: snapshoting data to file")
				storage.Save(counter)
			case <- quit:
				sticker.Stop()
				return
			}
		}
	}()
}


const DATA_PATH        = "data/counter.json"
const BIND_ADDR        = "0.0.0.0:6666"
const REQUEST_TTL      = 60
const REFRESH_INTERVAL = 1
const SAVE_INTERVAL    = 90

var counter *RequestCounter
var storage *RequestCounterStorage

var bind      string
var ttl       int
var path      string
var help      bool

var rInterval int = REFRESH_INTERVAL
var sInterval int = SAVE_INTERVAL

func main() {
	flag.StringVar(&bind, "bind", BIND_ADDR, "which address to bind to in the form of addr:port.")
	flag.IntVar(&ttl, "ttl", REQUEST_TTL, "request ttl in seconds. when requests expire they are no longer counted.")
	flag.StringVar(&path, "path", DATA_PATH, "path to the storage filename.")
	flag.BoolVar(&help, "help", false, "display this help text.")
	flag.Parse()

	if help {
		fmt.Printf("Usage: %s [-bind address] [-ttl ttl] [-path path]\n\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	counter = NewRequestCounter(time.Duration(ttl) * time.Second)
	storage = NewRequestCounterStorage(path)

	boot()
	start()
}
