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

func (t *ImageDedupThread) sendHTTPRequest(method, endpoint string, requestBody, response any) error {
	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建 URL
	url := t.baseURL + endpoint

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

		timeout := time.NewTimer(5 * time.Second)
		defer timeout.Stop()
		select {
		case <-done:
			log.Info().Msg("python API server process terminated gracefully")
		case <-timeout.C:
			log.Warn().Msg("python API server process did not terminate gracefully, force killing")
			_ = t.cmd.Process.Kill()
			<-done // 等待强制终止完成
		}
	}

	t.cmd = nil
	t.client = nil
	log.Info().Msg("HTTP API server thread closed")
}
