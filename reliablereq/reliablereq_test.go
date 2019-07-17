package reliablereq

import (
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/globocom/reliable-request/reliablereq"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func Test_It_returns_a_valid_response(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list").
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	// we need to intercept current http client due
	// https://github.com/h2non/gock/issues/27#issuecomment-334177773
	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	body, err := req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "{\"name\":\"mock\"}\n", body)
}

func Test_It_raises_an_error_when_there_is_no_connection(t *testing.T) {
	reliablereq.Flush()

	req := reliablereq.NewReliableRequest()

	_, err := req.Get("http://example.non/list")

	assert.NotNil(t, err)
}

func Test_It_returns_an_error_when_server_responds_with_a_non_2xx(t *testing.T) {

	defer gock.Off()
	reliablereq.Flush()

	non2xx := []int{400, 404, 503}
	for _, status := range non2xx {
		gock.New("http://example.com").
			Get("/list").
			Reply(status)

		req := reliablereq.NewReliableRequest()
		gock.InterceptClient(req.HTTPClient)
		defer gock.RestoreClient(req.HTTPClient)

		_, err := req.Get("http://example.com/list")

		assert.NotNil(t, err)
	}
}

func Test_It_uses_cache_when_enabled(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list").
		Times(1).
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	body, err := req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "{\"name\":\"mock\"}\n", body)

	body, err = req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "{\"name\":\"mock\"}\n", body)
}

func Test_It_doesnt_use_cache_when_disabled(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list").
		Times(1).
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	req.EnableCache = false
	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	body, err := req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "{\"name\":\"mock\"}\n", body)

	body, err = req.Get("http://example.com/list")

	assert.NotNil(t, err)
}

func Test_It_uses_stale_cache_when_enabled(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list").
		Times(1).
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	req.TTLCache = 1 * time.Second
	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	body, err := req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "{\"name\":\"mock\"}\n", body)

	// simulating a cache eviction
	time.Sleep(2 * time.Second)

	body, err = req.Get("http://example.com/list")

	assert.Nil(t, err)
}

func Test_It_doesnt_use_stale_cache_when_disabled(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list").
		Times(1).
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	req.TTLCache = 1 * time.Second
	req.EnableStaleCache = false
	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	body, err := req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "{\"name\":\"mock\"}\n", body)

	// simulating a cache eviction
	time.Sleep(2 * time.Second)

	body, err = req.Get("http://example.com/list")

	assert.NotNil(t, err)
}

func Test_It_allows_custom_headers(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		MatchHeader("Authorization", "^foo bar$").
		Get("/list").
		Times(1).
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	req.Headers = map[string]string{"Authorization": "foo bar"}

	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	body, err := req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "{\"name\":\"mock\"}\n", body)
}
func Test_It_opens_the_circuit_breaker_when_error_percentage_is_reached(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list").
		Times(5).
		Reply(503)

	req := reliablereq.NewReliableRequest()
	req.UpdateHystrixConfig("custom_cb", hystrix.CommandConfig{
		Timeout:                800 + 100, // the defaultTimeout http client + a small gap
		MaxConcurrentRequests:  100,
		ErrorPercentThreshold:  50,
		RequestVolumeThreshold: 4,
		SleepWindow:            5000,
	})

	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	body, _ := req.Get("http://example.com/list")
	body, _ = req.Get("http://example.com/list")
	body, _ = req.Get("http://example.com/list")
	body, _ = req.Get("http://example.com/list")
	body, _ = req.Get("http://example.com/list")

	cb, _, _ := hystrix.GetCircuit("custom_cb")

	assert.Equal(t, "", body)
	assert.Equal(t, true, cb.IsOpen())
}
func Test_It_closes_the_circuit_breaker_after_the_sleep_window(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list0").
		Reply(503)

	gock.New("http://example.com").
		Get("/list1").
		Reply(503)

	gock.New("http://example.com").
		Get("/list2").
		Reply(503)

	gock.New("http://example.com").
		Get("/list3").
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	gock.New("http://example.com").
		Get("/list4").
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	gock.New("http://example.com").
		Get("/list5").
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	req.UpdateHystrixConfig("custom_cb", hystrix.CommandConfig{
		Timeout:                800 + 100, // the defaultTimeout http client + a small gap
		MaxConcurrentRequests:  100,
		ErrorPercentThreshold:  50,
		RequestVolumeThreshold: 2,
		SleepWindow:            2000,
	})

	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	cb, _, _ := hystrix.GetCircuit("custom_cb")

	body, _ := req.Get("http://example.com/list0") //error 100%
	body, _ = req.Get("http://example.com/list1")  //error 100%

	assert.Equal(t, "", body)
	assert.Equal(t, true, cb.IsOpen()) // open due to error percentage + request volume threshold

	body, _ = req.Get("http://example.com/list2") //error 100%
	body, _ = req.Get("http://example.com/list3") //error 75%

	time.Sleep(3500 * time.Millisecond) // simulating its sleep window

	body, _ = req.Get("http://example.com/list4") //error 60%
	body, _ = req.Get("http://example.com/list5") //error 50%

	assert.Equal(t, false, cb.IsOpen()) // closed due to error percentage
}
