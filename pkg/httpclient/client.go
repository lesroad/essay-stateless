package httpclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout:   300 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func readResponseBodyForError(body io.ReadCloser, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 1024 // 默认最大1KB
	}

	content, err := io.ReadAll(io.LimitReader(body, int64(maxLength)))
	if err != nil {
		return fmt.Sprintf("failed to read response body: %v", err)
	}

	return string(content)
}

func (c *Client) Post(ctx context.Context, url string, data interface{}, result interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyContent := readResponseBodyForError(resp.Body, 1024)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, bodyContent)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *Client) PostWithHeaders(ctx context.Context, url string, data interface{}, result interface{}, headers map[string]string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyContent := readResponseBodyForError(resp.Body, 1024)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, bodyContent)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *Client) PostWithStream(ctx context.Context, url string, headers map[string]string, body interface{}, resultChan chan<- string) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyContent := readResponseBodyForError(resp.Body, 1024)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, bodyContent)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")

			var eventMap map[string]interface{}
			if err := json.Unmarshal([]byte(data), &eventMap); err != nil {
				logrus.Errorf("JSON解析错误: %v, 原始数据: %s", err, data)
				continue
			}

			if eventMap["type"] == "content" {
				resultChan <- data
			} else if eventMap["type"] == "end" {
				return nil
			} else {
				return fmt.Errorf("服务器错误: %s", data)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}
