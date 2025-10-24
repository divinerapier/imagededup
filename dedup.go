package imagededup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "embed"
)

var (
	//go:embed python_src/find_duplicates.py
	findDuplicatesScript string

	//go:embed python_src/find_duplicates_to_remove.py
	findDuplicatesToRemoveScript string
)

const (
	findDuplicatesScriptName         = "find_duplicates.py"
	findDuplicatesToRemoveScriptName = "find_duplicates_to_remove.py"
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

	Algorithm string
)

const (
	AlgorithmPHasher Algorithm = "phasher"
	AlgorithmCNN     Algorithm = "cnn"
	AlgorithmDHash   Algorithm = "dhash"
	AlgorithmAHash   Algorithm = "ahash"
	AlgorithmWHash   Algorithm = "whash"
)

// 将嵌入的脚本写入临时文件
func writeScriptToTemp(name, script string) (string, error) {
	tmpFile := os.TempDir() + "/" + name

	_, err := os.Stat(tmpFile)
	if err == nil {
		return tmpFile, nil
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	return tmpFile, os.WriteFile(tmpFile, []byte(script), 0755)
}

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

func FindDuplicates(algorithm Algorithm, imageDir string) (map[string]DuplicateResult, error) {
	// 将嵌入的脚本写入临时文件
	scriptPath, err := writeScriptToTemp(findDuplicatesScriptName, findDuplicatesScript)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp script: %w", err)
	}

	// 运行 Python 脚本
	output, err := runPythonScript(scriptPath, string(algorithm), imageDir)
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

func FindDuplicatesToRemove(algorithm Algorithm, imageDir string) ([]string, error) {
	// 将嵌入的脚本写入临时文件
	scriptPath, err := writeScriptToTemp(findDuplicatesToRemoveScriptName, findDuplicatesToRemoveScript)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp script: %w", err)
	}

	output, err := runPythonScript(scriptPath, string(algorithm), imageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run Python script: %w", err)
	}

	var result []string
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}
	return result, nil
}
