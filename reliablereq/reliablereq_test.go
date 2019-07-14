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

func TestCreateNewDefaultReliableRequest(t *testing.T) {
	req := reliablereq.NewReliableRequest()

	assert.NotNil(t, req)
	assert.Equal(t, req.HystrixCommandName, "default_config")
	assert.Equal(t, req.EnableCache, true)
	assert.Equal(t, req.EnableStaleCache, true)
}

func TestCreateCustomReliableRequest(t *testing.T) {
	req := reliablereq.ReliableRequest{
		HystrixCommandName: "custom_command",
		EnableStaleCache:   false,
		EnableCache:        false,
	}

	assert.NotNil(t, req)
	assert.Equal(t, req.HystrixCommandName, "custom_command")
	assert.Equal(t, req.EnableCache, false)
	assert.Equal(t, req.EnableStaleCache, false)
}

func TestValidRequest(t *testing.T) {
	defer gock.Off()
	reliablereq.FlushCache()

	gock.New("http://example.com").
		Get("/list").
		Reply(200).
		JSON(map[string]interface{}{"name": "mock"})

	req := reliablereq.NewReliableRequest()
	// we need to intercept current http client due
	// https://github.com/h2non/gock/issues/27#issuecomment-334177773<Paste>
	gock.InterceptClient(req.HTTPClient)
	defer gock.RestoreClient(req.HTTPClient)

	resp, err := req.Request("http://example.com/list")

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, string(body), "{\"name\":\"mock\"}\n")
}

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
