// vi:set tabstop=8 shiftwidth=8 noexpandtab:
package counter

import "time"

const (
	// default path + filename where to store the data in
	DefaultDataPath   = "./data/counter.json"

	// default bind address to listen to
	DefaultBindAddr   = "0.0.0.0:6666"

	// default request time to live (valid string for time.ParseDuration)
	DefaultRequestTtl = "60s"

	// max concurrent clients allowed
	MaxClients        = 5

	// how often should the counter data be refreshed (string for time.ParseDuration)
	RefreshInterval   = "1s"

	// how often should the counter data be saved to disk (string for time.ParseDuration)
	SaveInterval      = "90s"

	// for how long should each request sleep
	SleepPerRequest   = "2s"

	// prefix used for the log messages
	LoggerPrefix      = "go-request-counter"
)

type Config struct {
	BindAddr        string
	DataPath        string
	MaxClients      int
	RequestTtl      time.Duration
	RefreshInterval time.Duration
	SaveInterval    time.Duration
	SleepPerRequest time.Duration
}

func NewConfig(
	bindAddr string,
	dataPath string,
	maxClients int,
	requestTtl string,
	refreshInterval string,
	saveInterval string,
	sleepPerRequest string,
) (*Config, error) {
	_requestTtl, err := time.ParseDuration(requestTtl)
	if err != nil {
		return nil, err
	}
	_refreshInterval, err := time.ParseDuration(refreshInterval)
	if err != nil {
		return nil, err
	}
	_saveInterval, err := time.ParseDuration(saveInterval)
	if err != nil {
		return nil, err
	}
	_sleepPerRequest, err := time.ParseDuration(sleepPerRequest)
	if err != nil {
		return nil, err
	}

	return &Config{
		BindAddr:        bindAddr,
		DataPath:        dataPath,
		MaxClients:      maxClients,
		RequestTtl:      _requestTtl,
		RefreshInterval: _refreshInterval,
		SaveInterval:    _saveInterval,
		SleepPerRequest: _sleepPerRequest,
	}, nil
}
