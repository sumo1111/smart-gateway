package relay

import (
	"io"
	"net/http"
	"strings"
)

// OpenAIAdaptor OpenAI兼容适配器（适用于OpenAI/Azure/Custom等兼容端点）
type OpenAIAdaptor struct {
	BaseURL string
	APIKey  string
}

func (a *OpenAIAdaptor) ConvertRequest(body []byte) ([]byte, error) {
	// OpenAI格式直接透传，无需转换
	return body, nil
}

func (a *OpenAIAdaptor) ConvertResponse(resp *http.Response) ([]byte, error) {
	// OpenAI格式直接透传
	return io.ReadAll(resp.Body)
}

func (a *OpenAIAdaptor) GetBaseURL() string { return a.BaseURL }
func (a *OpenAIAdaptor) GetAPIKey() string  { return a.APIKey }

func (a *OpenAIAdaptor) DoRequest(url_, method string, headers map[string]string, body io.Reader) (*http.Response, error) {
	if headers == nil { headers = make(map[string]string) }
	headers["Authorization"] = "Bearer " + a.APIKey
	headers["Content-Type"] = "application/json"
	return DoHTTPRequest(url_, method, headers, body)
}

// RelayOpenAI 转发OpenAI兼容请求
func RelayOpenAI(baseURL, apiKey string, path string, method string, body io.Reader) (*http.Response, error) {
	url := strings.TrimRight(baseURL, "/") + path
	adaptor := &OpenAIAdaptor{BaseURL: baseURL, APIKey: apiKey}
	return adaptor.DoRequest(url, method, nil, body)
}
