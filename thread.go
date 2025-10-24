package imagededup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// HTTP 请求和响应结构
type FindDuplicatesRequest struct {
	Algorithm string `json:"algorithm"`
	ImageDir  string `json:"image_dir"`
}

type FindDuplicatesToRemoveRequest struct {
	Algorithm string `json:"algorithm"`
	ImageDir  string `json:"image_dir"`
}

type APIResponse[T any] struct {
	Status  string  `json:"status"`
	Data    T       `json:"data"`
	Message string  `json:"message"`
	Elapsed float64 `json:"elapsed"`
}

type ImageDedupThread struct {
	dedup *ImageDedup

	cmd        *exec.Cmd
	baseURL    string
	client     *http.Client
	serverPort int
}

func NewImageDedupThread(dedup *ImageDedup) (*ImageDedupThread, error) {
	// 选择一个可用端口
	port := 8000
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	thread := &ImageDedupThread{
		dedup:      dedup,
		baseURL:    baseURL,
		serverPort: port,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 启动 Python API 服务器
	if err := thread.startAPIServer(); err != nil {
		return nil, fmt.Errorf("failed to start API server: %w", err)
	}

	// 等待服务器启动并检查健康状态
	if err := thread.waitForServer(); err != nil {
		return nil, fmt.Errorf("server health check failed: %w", err)
	}

	log.Info().Str("baseURL", baseURL).Msg("API server started successfully")
	return thread, nil
}

func (t *ImageDedupThread) startAPIServer() error {
	// 构建启动命令
	args := []string{
		filepath.Join(t.dedup.source.GetExtractedPath(), "main.py"),
		"--port", fmt.Sprintf("%d", t.serverPort),
		"--host", "127.0.0.1",
	}

	log.Info().Str("args", strings.Join(args, " ")).Msg("starting python API server")

	// 使用嵌入的 Python 环境启动服务器
	cmd, err := t.dedup.engine.PythonCmd(args...)
	if err != nil {
		return err
	}

	// 将 stderr 重定向到日志
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	t.cmd = cmd

	// 启动进程
	if err := t.cmd.Start(); err != nil {
		return err
	}

	log.Info().Int("pid", t.cmd.Process.Pid).Msg("python API server process started")
	return nil
}

func (t *ImageDedupThread) waitForServer() error {
	maxRetries := 30
	retryInterval := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		if err := t.healthCheck(); err == nil {
			return nil
		}

		log.Info().Int("attempt", i+1).Int("max_retries", maxRetries).Msg("waiting for server to be ready")
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("server failed to start after %d attempts", maxRetries)
}

func (t *ImageDedupThread) healthCheck() error {
	resp, err := t.client.Get(t.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (t *ImageDedupThread) executeFindDuplicates(algorithm Algorithm, imageDir string) ([]DuplicateResult, error) {
	log.Info().Str("algorithm", string(algorithm)).Str("image_dir", imageDir).Msg("executing find duplicates via HTTP API")

	// 构建请求
	req := FindDuplicatesRequest{
		Algorithm: string(algorithm),
		ImageDir:  imageDir,
	}

	// 发送 HTTP 请求
	var response APIResponse[map[string]DuplicateResult]
	if err := t.sendHTTPRequest("POST", "/find-duplicates", req, &response); err != nil {
		log.Error().Err(err).Msg("error executing find duplicates via HTTP")
		return nil, err
	}

	if response.Status != "ok" {
		return nil, errors.New(response.Message)
	}

	// 将 map 转换为 slice
	var results []DuplicateResult
	for _, result := range response.Data {
		results = append(results, result)
	}

	return results, nil
}

func (t *ImageDedupThread) executeFindDuplicatesToRemove(algorithm Algorithm, imageDir string) ([]string, error) {
	log.Info().Str("algorithm", string(algorithm)).Str("image_dir", imageDir).Msg("executing find duplicates to remove via HTTP API")

	// 构建请求
	req := FindDuplicatesToRemoveRequest{
		Algorithm: string(algorithm),
		ImageDir:  imageDir,
	}

	// 发送 HTTP 请求
	var response APIResponse[[]string]
	if err := t.sendHTTPRequest("POST", "/find-duplicates-to-remove", req, &response); err != nil {
		log.Error().Err(err).Msg("error executing find duplicates to remove via HTTP")
		return nil, err
	}

	if response.Status != "ok" {
		return nil, errors.New(response.Message)
	}

	return response.Data, nil
}

func (t *ImageDedupThread) sendHTTPRequest(method, endpoint string, requestBody interface{}, response interface{}) error {
	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建 URL
	url := t.baseURL + endpoint
	// log.Info().Str("method", method).Str("url", url).Str("body", string(jsonData)).Msg("sending HTTP request")

	// 创建 HTTP 请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// log.Info().Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("received HTTP response")

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (t *ImageDedupThread) Close() {
	log.Info().Msg("closing HTTP API server thread")

	if t.cmd != nil && t.cmd.Process != nil {
		log.Info().Int("pid", t.cmd.Process.Pid).Msg("terminating python API server process")

		// 优雅关闭：发送 SIGTERM
		if err := t.cmd.Process.Signal(os.Interrupt); err != nil {
			log.Warn().Err(err).Msg("failed to send SIGTERM, will force kill")
		}

		// 等待进程结束，最多等待 5 秒
		done := make(chan error, 1)
		go func() {
			done <- t.cmd.Wait()
		}()

		select {
		case <-done:
			log.Info().Msg("python API server process terminated gracefully")
		case <-time.After(5 * time.Second):
			log.Warn().Msg("python API server process did not terminate gracefully, force killing")
			_ = t.cmd.Process.Kill()
			<-done // 等待强制终止完成
		}
	}

	t.cmd = nil
	t.client = nil
	log.Info().Msg("HTTP API server thread closed")
}
