package imagededup

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/divinerapier/go-embed-python/embed_util"
	"github.com/divinerapier/go-embed-python/python"
	"github.com/divinerapier/imagededup/internal/data"
	"github.com/divinerapier/imagededup/python_src"
	"github.com/rs/zerolog/log"
)

type ImageDedup struct {
	engine  python.Python
	library *embed_util.EmbeddedFiles
	source  *embed_util.EmbeddedFiles

	// HTTP 客户端配置
	baseURL    string
	client     *http.Client
	serverPort int

	// 线程池配置
	parallelism int
	threads     chan *ImageDedupThread

	// 服务器管理
	serverProcess *os.Process
	mu            sync.RWMutex
}

func NewImageDedup(parallelism int) (*ImageDedup, error) {
	tmpDir := filepath.Join(os.TempDir(), "go-imagededup-embedded")
	engine, err := python.NewEmbeddedPythonWithTmpDir(tmpDir+"-python", true)
	if err != nil {
		return nil, err
	}

	library, err := embed_util.NewEmbeddedFilesWithTmpDir(data.Data, tmpDir+"-imagededup-lib", true)
	if err != nil {
		return nil, err
	}

	engine.AddPythonPath(library.GetExtractedPath())

	source, err := embed_util.NewEmbeddedFilesWithTmpDir(python_src.Scripts, tmpDir+"-imagededup-source", true)
	if err != nil {
		return nil, err
	}

	// 选择一个可用端口
	serverPort := 18000
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", serverPort)

	dedup := &ImageDedup{
		engine:      engine,
		library:     library,
		source:      source,
		baseURL:     baseURL,
		serverPort:  serverPort,
		parallelism: parallelism,
		threads:     make(chan *ImageDedupThread, parallelism),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 启动 HTTP 服务器
	if err := dedup.startHTTPServer(); err != nil {
		dedup.Cleanup()
		return nil, fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// 等待服务器启动
	if err := dedup.waitForServer(); err != nil {
		dedup.Cleanup()
		return nil, fmt.Errorf("server health check failed: %w", err)
	}

	// 初始化线程池
	if err := dedup.initializeThreadPool(); err != nil {
		dedup.Cleanup()
		return nil, fmt.Errorf("failed to initialize thread pool: %w", err)
	}

	log.Info().Str("baseURL", baseURL).Int("parallelism", parallelism).Msg("ImageDedup initialized successfully")
	return dedup, nil
}

// 启动 HTTP 服务器
func (i *ImageDedup) startHTTPServer() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 构建启动命令
	args := []string{
		filepath.Join(i.source.GetExtractedPath(), "main.py"),
		"--port", fmt.Sprintf("%d", i.serverPort),
		"--host", "127.0.0.1",
	}

	log.Info().Strs("args", args).Msg("starting python HTTP server")

	// 使用嵌入的 Python 环境启动服务器
	cmd, err := i.engine.PythonCmd(args...)
	if err != nil {
		return fmt.Errorf("failed to create python command: %w", err)
	}

	// 将 stderr 重定向到日志
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start python server: %w", err)
	}

	i.serverProcess = cmd.Process
	log.Info().Int("pid", cmd.Process.Pid).Msg("python HTTP server process started")
	return nil
}

// 等待服务器启动并检查健康状态
func (i *ImageDedup) waitForServer() error {
	maxRetries := 300
	retryInterval := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := i.healthCheck(); err == nil {
			return nil
		}

		log.Info().
			Str("baseURL", i.baseURL).
			Int("attempt", attempt+1).
			Int("max_retries", maxRetries).
			Msg("waiting for server to be ready")
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("server failed to start after %d attempts", maxRetries)
}

// 健康检查
func (i *ImageDedup) healthCheck() error {
	resp, err := i.client.Get(i.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// 初始化线程池
func (i *ImageDedup) initializeThreadPool() error {
	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
		err   error
	)

	// 创建线程池
	for index := 0; index < i.parallelism; index++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			log.Info().Int("index", index).Msg("creating HTTP client thread")

			// 创建 HTTP 客户端线程（不需要启动单独的进程）
			thread := &ImageDedupThread{
				dedup:      i,
				baseURL:    i.baseURL,
				client:     i.client,
				serverPort: i.serverPort,
			}

			mutex.Lock()
			if err == nil {
				i.threads <- thread
				log.Info().Int("index", index).Msg("HTTP client thread created")
			}
			mutex.Unlock()
		}(index)
	}

	log.Info().Msg("waiting for thread pool initialization")
	wg.Wait()
	log.Info().Msg("thread pool initialization completed")

	return err
}

func (i *ImageDedup) Cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 关闭 HTTP 服务器
	if i.serverProcess != nil {
		log.Info().Int("pid", i.serverProcess.Pid).Msg("terminating HTTP server process")

		var wg sync.WaitGroup
		wg.Add(1)
		// 等待进程结束，最多等待 5 秒
		done := make(chan error, 1)
		go func() {
			wg.Done()
			_, err := i.serverProcess.Wait()
			done <- err
		}()

		wg.Wait()

		// 优雅关闭：发送 SIGTERM
		if err := i.serverProcess.Signal(os.Interrupt); err != nil {
			log.Warn().Err(err).Msg("failed to send SIGTERM, will force kill")
		}

		select {
		case <-done:
			log.Info().Msg("HTTP server process terminated gracefully")
		case <-time.After(5 * time.Second):
			log.Warn().Msg("HTTP server process did not terminate gracefully, force killing")
			_ = i.serverProcess.Kill()
			<-done // 等待强制终止完成
		}

		i.serverProcess = nil
	}

	// 清理资源
	i.library.Cleanup()
	i.source.Cleanup()

	if engine, ok := i.engine.(*python.EmbeddedPython); ok {
		engine.Cleanup()
	}

	log.Info().Msg("ImageDedup cleanup completed")
}

func (i *ImageDedup) Close() {
	log.Info().Msg("closing ImageDedup")

	close(i.threads)

	// 关闭线程池
	for range i.threads {
	}

	// 关闭 HTTP 服务器
	i.Cleanup()

	log.Info().Msg("all HTTP client threads closed")
}

func (i *ImageDedup) FindDuplicates(algorithm Algorithm, imageDir string) ([]DuplicateResult, error) {
	log.Info().Str("algorithm", string(algorithm)).Str("image_dir", imageDir).Msg("finding duplicates via HTTP API")

	// 获取线程池中的 HTTP 客户端
	thread, ok := <-i.threads
	if !ok {
		return nil, errors.New("no HTTP client threads available")
	}
	defer func() {
		log.Info().Msg("returning HTTP client thread to pool")
		i.threads <- thread
	}()

	// 执行 HTTP 请求
	results, err := thread.executeFindDuplicates(algorithm, imageDir)
	if err != nil {
		log.Error().Err(err).Msg("error executing find duplicates via HTTP")
		return nil, err
	}

	log.Info().Int("count", len(results)).Msg("find duplicates completed successfully")
	return results, nil
}

func (i *ImageDedup) FindDuplicatesToRemove(algorithm Algorithm, imageDir string) ([]string, error) {
	log.Info().Str("algorithm", string(algorithm)).Str("image_dir", imageDir).Msg("finding duplicates to remove via HTTP API")

	// 获取线程池中的 HTTP 客户端
	thread, ok := <-i.threads
	if !ok {
		return nil, errors.New("no HTTP client threads available")
	}
	defer func() {
		log.Info().Msg("returning HTTP client thread to pool")
		i.threads <- thread
	}()

	// 执行 HTTP 请求
	results, err := thread.executeFindDuplicatesToRemove(algorithm, imageDir)
	if err != nil {
		log.Error().Err(err).Msg("error executing find duplicates to remove via HTTP")
		return nil, err
	}

	log.Info().Int("count", len(results)).Msg("find duplicates to remove completed successfully")
	return results, nil
}
