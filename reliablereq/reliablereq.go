package reliablereq

import (
	"fmt"
	"io/ioutil"
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

func Flush() {
	cachedResponses.Flush()
	hystrix.Flush()
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

// Get - returns the requested data as string and a possible error
func (rr *ReliableRequest) Get(url string) (string, error) {
	var body string

	if rr.EnableCache {
		cached, found := rr.getCache(url)
		if found {
			return cached, nil
		}
	}

	err := hystrix.Do(rr.HystrixCommandName, func() error {
		resp, err := rr.urlRequest(url)
		if err == nil {
			rawBody, _ := ioutil.ReadAll(resp.Body)
			body = string(rawBody)
			rr.setCache(url, body)
		}
		return err
	}, func(previousError error) error {
		var err error

		if rr.EnableStaleCache {
			body, err = rr.getStaleCache(url)
			if err != nil {
				return errors.Wrap(previousError, err.Error())
			}
		} else {
			return previousError
		}

		return nil
	})

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not complete the request for url"))
	}

	return body, nil
}

func keyStale(key string) string {
	return fmt.Sprintf("%s-stale", key)
}

func (rr *ReliableRequest) setCache(url, body string) {
	cachedResponses.Set(url, body, rr.TTLCache)

	if rr.EnableCache && rr.EnableStaleCache {
		cachedResponses.Set(keyStale(url), body, rr.TTLStaleCache)
	}
}

func (rr *ReliableRequest) getCache(key string) (string, bool) {
	cached, found := cachedResponses.Get(key)
	if found {
		return cached.(string), true
	}

	return "", false
}

func (rr *ReliableRequest) getStaleCache(key string) (string, error) {
	cached, found := rr.getCache(keyStale(key))
	if found {
		return cached, nil
	}
	return "", errors.New("failed to fetch stale response")
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
