// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package main

import (
	"io"
	"fmt"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/op/go-logging"

	"github.com/nocive/go-request-counter/src/counter"
	"github.com/nocive/go-request-counter/src/storage"
)

func handleError(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		logger.Fatal(fmt.Sprintf("Line: %d\tFile: %s\nMessage: %s", line, file, err))
		os.Exit(1)
	}
}

func boot() {
	var err error

	logger.Info(fmt.Sprintf("booting! ttl: %dms / refresh interval: %dms / save interval: %dms\n", requestTtl, RefreshInterval, SaveInterval))

	if !reqCounterStorage.Exists() {
		logger.Info("data file doesn't exist, creating\n")
		err = reqCounterStorage.Create()
		handleError(err)
	} else {
		logger.Info("data file exists, loading\n")
		err = reqCounterStorage.Load(reqCounter)
		handleError(err)
		logger.Info(fmt.Sprintf("%d requests loaded from data file\n", reqCounter.GetCount()))

		logger.Info("firing a manual refresh\n")
		reqCounter.Refresh() // purge expired events before starting
	}

	logger.Info("initializing signal traps\n")
	trap()

	logger.Info("initializing tickers\n")
	ticker()
}

func start() {
	sema := make(chan struct{}, MaxClients)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()
		process(w, r)
	})

	logger.Info(fmt.Sprintf("preparing to listen on %s\n", bindAddr))
	http.ListenAndServe(bindAddr, nil)
}

func process(w http.ResponseWriter, r *http.Request) {
	remoteAddr := r.Header.Get("X-Client-IP")
	if remoteAddr == "" {
		remoteAddr = r.RemoteAddr
	}
	logger.Info(fmt.Sprintf("request received / ip: %s\n", remoteAddr))

	reqCounter.Mark(remoteAddr)
	currentCount := reqCounter.GetCount()

	io.WriteString(w, fmt.Sprintf("%d\n", currentCount))
	logger.Info(fmt.Sprintf("hits incremented / current: %d\n", currentCount))

	time.Sleep(SleepPerRequest * time.Millisecond)
	logger.Info("request finished\n")
}

func shutdown() {
	logger.Info("saving data to file\n")

	err := reqCounterStorage.Save(reqCounter)
	handleError(err)
}

func trap() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Info("caught signal, cleaning up...\n")
		shutdown()
		os.Exit(0)
	}()
}

func ticker() {
	quit := make(chan struct{})

	refreshTicker := time.NewTicker(time.Duration(RefreshInterval) * time.Millisecond)
	saveTicker := time.NewTicker(time.Duration(SaveInterval) * time.Millisecond)

	go func() {
		for {
			select {
			case <- refreshTicker.C:
				logger.Info(fmt.Sprintf("refresh / count: %d\n", reqCounter.GetCount()))
				reqCounter.Refresh()

			case <- saveTicker.C:
				logger.Info("snapshotting data to file\n")
				reqCounterStorage.Save(reqCounter)

			case <- quit:
				refreshTicker.Stop()
				saveTicker.Stop()
				return
			}
		}
	}()
}




const (
	// default path + filename where to store the data in
	DefaultDataPath   = "./data/counter.json"

	// default bind address to listen to
	DefaultBindAddr   = "0.0.0.0:6666"

	// default request time to live (ms)
	DefaultRequestTtl = 60000

	// max concurrent clients allowed
	MaxClients        = 5

	// how often should the counter data be refreshed (ms)
	RefreshInterval   = 1000

	// how often should the counter data be saved to disk (ms)
	SaveInterval      = 90000

	// for how long should each request sleep (ms)
	SleepPerRequest   = 2000

	// prefix used for the log messages
	LoggerPrefix      = "go-request-counter"
)

var logger = logging.MustGetLogger(LoggerPrefix)

var format = logging.MustStringFormatter(
	`%{color}• %{shortfunc} %{level:.4s} %{id:03x}%{color:reset} ‣ %{message}`,
)

var (
	bindAddr    string
	requestTtl  int
	dataPath    string
	displayHelp bool

	reqCounter        *counter.RequestCounter
	reqCounterStorage *storage.RequestCounterStorage
)

func main() {
	logging.SetFormatter(format)

	flag.StringVar(&bindAddr, "bind", DefaultBindAddr, "which address to bind to in the form of addr:port.")
	flag.IntVar(&requestTtl, "ttl", DefaultRequestTtl, "request ttl in seconds. when requests expire they are no longer counted.")
	flag.StringVar(&dataPath, "path", DefaultDataPath, "path to the storage filename.")
	flag.BoolVar(&displayHelp, "help", false, "display this help text.")
	flag.Parse()

	if displayHelp {
		fmt.Printf("Usage: %s [-bind address] [-ttl ttl] [-path path]\n\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	reqCounter = counter.NewRequestCounter(time.Duration(requestTtl) * time.Millisecond)
	reqCounterStorage = storage.NewRequestCounterStorage(dataPath)

	boot()
	start()
}
