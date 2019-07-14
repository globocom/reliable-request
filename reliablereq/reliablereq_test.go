package reliablereq

import (
	"github.com/globocom/reliable-request/reliablereq"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
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

	resp, err := req.Request("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(body), "{\"name\":\"mock\"}\n")
}

func Test_It_raises_error_when_there_is_no_connection(t *testing.T) {
	reliablereq.Flush()

	req := reliablereq.NewReliableRequest()

	resp, err := req.Request("http://example.non/list")

	assert.Nil(t, resp)
	assert.NotNil(t, err)
}

func Test_It_returns_error_when_server_responds_with_a_non_2xx(t *testing.T) {

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

		resp, err := req.Request("http://example.com/list")

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	}
}

// ############### Caching ######################

func Test_It_uses_cache_when_enabled(t *testing.T) {
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

	resp, err := req.Request("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(body), "{\"name\":\"mock\"}\n")

	resp, err = req.Request("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	//https://medium.com/@xoen/golang-read-from-an-io-readwriter-without-loosing-its-content-2c6911805361
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(body), "{\"name\":\"mock\"}\n")

}
func Test_It_doesnt_use_cache_when_disable(t *testing.T) {}

func Test_It_uses_stale_cache_when_enabled(t *testing.T)       {}
func Test_It_doesnt_use_stale_cache_when_disable(t *testing.T) {}

// ############### Caching ######################

// AllNon2xxAreError
//

// happy path
// 1. 200 com conteudo
// 2. caching (enabled/disabled)
// 3. stale caching (enabled/disabled)
// 4. custom headers

// unhappy path
// 1. non 2xx (should raise err)
// 2. no stale

// future
// [ ] retry policy
// [ ] async hystrix (Go instead of Do)
// [ ] load stress
