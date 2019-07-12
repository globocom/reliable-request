package reliablereq

import (
	"github.com/globocom/reliable-request/reliablereq"
	"github.com/stretchr/testify/assert"
	//	gock "gopkg.in/h2non/gock.v1"
	//	"net"
	//	"net/http"
	"testing"
	//	"time"
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

//func TestValidRequest(t *testing.T) {
//	defer gock.Off()
//
//	gock.New("http://example.com").
//		Get("/list").
//		Reply(200).
//		JSON([]map[string]interface{}{
//			{
//				"name":   "mock",
//				"number": 42,
//			},
//		})
//
//	req := reliablereq.NewReliableRequest()
//	resp, err := req.Request("http://example.com/list")
//
//	assert.Nil(t, err)
//	assert.NotNil(t, resp)
//}
//
//func TestSimple(t *testing.T) {
//	defer gock.Off()
//
//	gock.New("http://example.com").
//		Get("/bar.json").
//		Reply(200).
//		JSON(map[string]string{"foo": "bar"})
//
//	//client := &http.Client{}
//	client := &http.Client{
//		Transport: &http.Transport{
//			DialContext: (&net.Dialer{
//				Timeout:   1000,
//				KeepAlive: 30 * time.Second,
//			}).DialContext,
//			MaxIdleConns:        300,
//			MaxIdleConnsPerHost: 300,
//			TLSHandshakeTimeout: 300,
//		},
//		Timeout: 300,
//	}
//
//	req, err := http.NewRequest("GET", "http://example.com/bar.json", nil)
//	res, err := client.Do(req)
//	//res, err := http.Get("http://example.com/bar.json")
//	assert.Nil(t, err)
//	assert.Equal(t, res.StatusCode, 200)
//
//	//	body, _ := ioutil.ReadAll(res.Body)
//	//	st.Expect(t, string(body)[:13], `{"foo":"bar"}`)
//
//	//	st.Expect(t, gock.IsDone(), true)
//}

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
