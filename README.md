# Reliablereq
A golang opinionated library to provide reliable request using [hystrix-go](https://github.com/afex/hystrix-go), [go-cache](https://github.com/patrickmn/go-cache), and [go-resiliency](https://github.com/eapache/go-resiliency).

When you do a `Get`, it provides:
* an HTTP client configured with timeouts and [keepalive](https://en.wikipedia.org/wiki/HTTP_persistent_connection),
* a [circuit breaker](https://martinfowler.com/bliki/CircuitBreaker.html) and
* a proper [caching system](https://en.wikipedia.org/wiki/Cache_(computing)).

# Usage

```golang
req := reliablereq.NewReliableRequest()
req.TTLCache = 1 * time.Second
req.EnableStaleCache = false
body, err := req.Get("http://example.com/list")

// passing authentication/authorization bearer token
req := reliablereq.NewReliableRequest()
req.Headers = map[string]string{"Authorization": "Bearer foobar"}
body, err := req.Get("http://example.com/list")
```

# Opinionated defaults

```golang
// reliable request defaults
rr := ReliableRequest{
  EnableCache:        true,
  TTLCache:           1 * time.Minute,
  EnableStaleCache:   true,
  TTLStaleCache:      24 * time.Hour,
}
// hystrix
var defaultHystrixConfiguration = hystrix.CommandConfig{
  Timeout:                800 + 100, // the defaultTimeout http client + a small gap
  MaxConcurrentRequests:  100,
  ErrorPercentThreshold:  50,
  RequestVolumeThreshold: 3,
  SleepWindow:            5000,
}
// http client
client := &http.Client{
  Transport: &http.Transport{
    DialContext: (&net.Dialer{
      Timeout:   800 * time.Millisecond,
      KeepAlive: 30 * time.Second,
    }).DialContext,
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 100,
    TLSHandshakeTimeout: 800 * time.Millisecond,
  },
  Timeout: 800 * time.Millisecond,
}
```

# Future

* provide a proxy to setup hystrix
* add more examples, like token header requests and more
* discuss the adopted defaults
* discuss whether async hystrix is better (Go instead of Do)
* understand and test the simultaneous client req hystrix config to see its implications
* add travis ci
* add go api documentation
* add retry logic (by go-resiliency)
* add hooks (callbacks) to provides means for metrics gathering
* add more HTTP verbs?
* add load stress
