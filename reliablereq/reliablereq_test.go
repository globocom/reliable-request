package reliablereq

import (
	"github.com/globocom/reliable-request/reliablereq"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
	"testing"
	"time"
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
		// we need to intercept current http client due
		// https://github.com/h2non/gock/issues/27#issuecomment-334177773
		gock.InterceptClient(req.HTTPClient)
		defer gock.RestoreClient(req.HTTPClient)

		_, err := req.Get("http://example.com/list")

		assert.NotNil(t, err)
	}
}

// ############### Caching ######################
func Test_It_uses_cache_when_enabled(t *testing.T) {
	defer gock.Off()
	reliablereq.Flush()

	gock.New("http://example.com").
		Get("/list").
		Times(1).
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
	// we need to intercept current http client due
	// https://github.com/h2non/gock/issues/27#issuecomment-334177773
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
	// we need to intercept current http client due
	// https://github.com/h2non/gock/issues/27#issuecomment-334177773
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
	// we need to intercept current http client due
	// https://github.com/h2non/gock/issues/27#issuecomment-334177773
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

// ############### Caching ######################

// AllNon2xxAreError
//

// happy path
// 1. 200 com conteudo
// 2. caching (enabled/disabled)
// 3. stale caching (enabled/disabled)
// 4. custom headers
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
