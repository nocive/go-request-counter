// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package main

import (
	"fmt"
	"flag"
	"os"

	"github.com/op/go-logging"

	"github.com/nocive/go-request-counter/src/counter"
	"github.com/nocive/go-request-counter/src/storage"
	"github.com/nocive/go-request-counter/src/request"
)

var logger = logging.MustGetLogger(counter.LoggerPrefix)

var format = logging.MustStringFormatter(
	`%{color}• %{shortfunc} %{level:.4s} %{id:03x}%{color:reset} ‣ %{message}`,
)

var (
	bindAddr        string
	dataPath        string
	maxClients      int    = counter.MaxClients
	requestTtl      string
	refreshInterval string = counter.RefreshInterval
	saveInterval    string = counter.SaveInterval
	sleepPerRequest string = counter.SleepPerRequest

	displayHelp     bool

	cfg *counter.Config
	cnt *counter.RequestCounter
	bck *request.RequestBucket
	stg *storage.RequestCounterStorage
)

func main() {
	logging.SetFormatter(format)

	flag.StringVar(&bindAddr, "bind", counter.DefaultBindAddr, "which address to bind to in the form of addr:port.")
	flag.StringVar(&requestTtl, "ttl", counter.DefaultRequestTtl, "request ttl expressed as a time duration string (eg: 30s for 30 seconds).")
	flag.StringVar(&dataPath, "path", counter.DefaultDataPath, "path to the storage filename.")
	flag.BoolVar(&displayHelp, "help", false, "display this help text.")
	flag.Parse()

	if displayHelp {
		fmt.Printf("Usage: %s [-bind address] [-ttl ttl] [-path path]\n\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	cfg, err := counter.NewConfig(
		bindAddr,
		dataPath,
		maxClients,
		requestTtl,
		refreshInterval,
		saveInterval,
		sleepPerRequest,
	)
	if err != nil {
		logger.Fatal(err)
	}

	bck = request.NewRequestBucket(cfg.RequestTtl)
	stg = storage.NewRequestCounterStorage(cfg.DataPath)
	cnt = counter.NewRequestCounter(*cfg, *bck, *stg, *logger)

	if err := cnt.Init(); err != nil {
		logger.Fatal(err)
	}
	cnt.Start()
}
