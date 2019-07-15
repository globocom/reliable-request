package reliablereq

import (
	"testing"
	"time"

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
	assert.Equal(t, body, "{\"name\":\"mock\"}\n")
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
	assert.Equal(t, body, "{\"name\":\"mock\"}\n")

	body, err = req.Get("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, body, "{\"name\":\"mock\"}\n")
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
	assert.Equal(t, body, "{\"name\":\"mock\"}\n")

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
	assert.Equal(t, body, "{\"name\":\"mock\"}\n")

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
	assert.Equal(t, body, "{\"name\":\"mock\"}\n")

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
	assert.Equal(t, body, "{\"name\":\"mock\"}\n")
}

// happy path
// 5. all hystrix features (circuit break on and off)

// unhappy path
// 1. non 2xx (should raise err)
// 2. no stale

// future
// [ ] retry policy
// [ ] async hystrix (Go instead of Do)
// [ ] understand and test the simultaneous client req hystrix config to see its implications
// [ ] load stress
// [ ] other reliable methods PUT POST
