package relay

import (
	"io"
	"net/http"
)

// Adaptor Provider适配器接口
type Adaptor interface {
	// ConvertRequest 将OpenAI格式请求转换为Provider原生格式
	ConvertRequest(body []byte) ([]byte, error)
	// ConvertResponse 将Provider响应转换回OpenAI格式
	ConvertResponse(resp *http.Response) ([]byte, error)
	// GetBaseURL 获取Provider的API地址
	GetBaseURL() string
	// GetAPIKey 获取API密钥
	GetAPIKey() string
	// DoRequest 执行实际请求
	DoRequest(url, method string, headers map[string]string, body io.Reader) (*http.Response, error)
}

// DoHTTPRequest 通用HTTP请求执行
func DoHTTPRequest(url, method string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil { return nil, err }
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	return client.Do(req)
}
