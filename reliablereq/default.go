package reliablereq

import (
	"net"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

const defaultTCPConnectionTimeout = 500 * time.Millisecond
const defaultTimeout = 800 * time.Millisecond
const defaultKeepAliveConnections = 100
const defaultCommandName = "default_config"

func timeoutHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   defaultTCPConnectionTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        defaultKeepAliveConnections,
			MaxIdleConnsPerHost: defaultKeepAliveConnections,
			TLSHandshakeTimeout: defaultTCPConnectionTimeout,
		},
		Timeout: defaultTimeout,
	}
	return client
}

func makeDefaultHistryx() {
	var defaultHystrixConfiguration = hystrix.CommandConfig{
		Timeout:                800 + 100, // the defaultTimeout http client + a small gap
		MaxConcurrentRequests:  100,
		ErrorPercentThreshold:  50,
		RequestVolumeThreshold: 3,
		SleepWindow:            5000,
	}

	hystrix.ConfigureCommand(defaultCommandName, defaultHystrixConfiguration)
}

func defaultReliableRequest() ReliableRequest {
	rr := ReliableRequest{
		HTTPClient:         timeoutHTTPClient(),
		EnableCache:        true,
		TTLCache:           1 * time.Minute,
		EnableStaleCache:   true,
		TTLStaleCache:      24 * time.Hour,
		hystrixCommandName: defaultCommandName,
	}
	return rr
}
