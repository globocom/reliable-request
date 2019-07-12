package reliablereq

import (
	"fmt"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

var cachedResponses *cache.Cache

func init() {
	makeDefaultHistryx()
	cachedResponses = cache.New(5*time.Minute, 10*time.Minute)
}

// ReliableRequest - a struct holding params to make reliable requests
type ReliableRequest struct {
	Headers            map[string]string
	HTTPClient         *http.Client
	EnableCache        bool
	TTLCache           time.Duration
	EnableStaleCache   bool
	TTLStaleCache      time.Duration
	HystrixCommandName string
	HystrixCommand     hystrix.CommandConfig
}

// NewReliableRequest - create a new ReliableRequest
func NewReliableRequest() *ReliableRequest {
	rr := defaultReliableRequest()

	return &rr
}

// Request - method to request data from a url
func (rr *ReliableRequest) Request(url string) (*http.Response, error) {
	var resp *http.Response

	if rr.EnableCache {
		cached, found := rr.getCache(url)
		if found {
			return cached, nil
		}
	}

	err := hystrix.Do(rr.HystrixCommandName, func() error {
		var err error
		resp, err = rr.urlRequest(url)
		if err == nil {
			rr.setCache(url, resp)
		}
		return err
	}, func(previousError error) error {
		var err error

		if rr.EnableStaleCache {
			resp, err = rr.getStaleCache(url)
			if err != nil {
				return errors.Wrap(previousError, err.Error())
			}
		} else {
			return previousError
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not complete the request for url"))
	}

	return resp, nil
}

func keyStale(key string) string {
	return fmt.Sprintf("%s-stale", key)
}

func (rr *ReliableRequest) setCache(url string, resp *http.Response) {
	cachedResponses.Set(url, resp, rr.TTLCache)

	if rr.EnableStaleCache {
		cachedResponses.Set(keyStale(url), resp, rr.TTLStaleCache)
	}
}

func (rr *ReliableRequest) getCache(key string) (*http.Response, bool) {
	cached, found := cachedResponses.Get(key)
	if found {
		return cached.(*http.Response), true
	}

	return nil, false
}

func (rr *ReliableRequest) getStaleCache(key string) (*http.Response, error) {
	cached, found := rr.getCache(keyStale(key))
	if found {
		return cached, nil
	}
	return nil, errors.New("failed to fetch stale response")
}

func (rr *ReliableRequest) urlRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not create an http request for url: %s", url))
	}

	if rr.Headers != nil {
		for k, v := range rr.Headers {
			req.Header.Add(k, v)
		}
	}
	resp, err := rr.HTTPClient.Do(req)

	if err == nil {
		// we only cache 2xx
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(fmt.Sprintf("bad response: %s for url: %s", resp.Status, url))
		}
	}

	return resp, err
}
