package imagededup

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type (
	DuplicateResult struct {
		Filename      string          `json:"filename"`
		DuplicateList []DuplicateItem `json:"duplicate_list"`
	}

	DuplicateItem struct {
		Filename string  `json:"filename"`
		Score    float64 `json:"score"`
	}
)

func runPythonScript(scriptPath string, args ...string) ([]byte, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("python %s %s", scriptPath, strings.Join(args, " ")))

	// 分别捕获 stdout 和 stderr
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// 如果命令执行失败，返回 stderr 作为错误
	if err != nil {
		return nil, fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	// 成功时返回 stdout 内容
	return []byte(stdout.String()), nil
}

func FindDuplicates(algorithm, imageDir string) (map[string]DuplicateResult, error) {
	// 运行 Python 脚本
	output, err := runPythonScript("find_duplicates.py", algorithm, imageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run Python script: %w", err)
	}

	// 解析 JSON 输出
	var result map[string]DuplicateResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return result, nil
}

func FindDuplicatesToRemove(algorithm, imageDir string) ([]string, error) {
	output, err := runPythonScript("find_duplicates_to_remove.py", algorithm, imageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run Python script: %w", err)
	}

	var result []string
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}
	return result, nil
}
