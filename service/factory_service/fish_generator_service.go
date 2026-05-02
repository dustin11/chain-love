package factory_service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"senspace/pkg/app/security"
)

const fishGeneratorPluginId = "FishTank"

// GenerateFishData 执行 FishTank 鱼数据生成器。
func GenerateFishData(user security.JwtUser, pluginIdRaw string, req GenerateFishDataRequest) (*GenerateFishDataResponse, error) {
	if user.Id == 0 {
		return nil, newParameterError("用户ID不能为空")
	}
	pluginId := strings.TrimSpace(pluginIdRaw)
	if pluginId != fishGeneratorPluginId {
		return nil, newParameterError("当前仅支持 FishTank 生成器")
	}

	mode := req.Mode
	if mode == "" {
		mode = GenerateFishDataModeTest
	}
	if mode != GenerateFishDataModeTest && mode != GenerateFishDataModeFormal {
		return nil, newParameterError("未知生成模式")
	}

	generatorDir, err := fishGeneratorDir()
	if err != nil {
		return nil, err
	}
	outputDir := filepath.Join(generatorDir, "generated")
	fishDirName := "fish-test"
	if mode == GenerateFishDataModeFormal {
		fishDirName = "fish"
	}

	args := []string{
		filepath.Join("dist", "cli", "generate.js"),
		"--output-dir", outputDir,
		"--fish-dir-name", fishDirName,
	}
	if mode == GenerateFishDataModeTest {
		count := normalizeFishGenerateCount(req.Count)
		args = append(args, "--limit-per-tier", strconv.Itoa(count))
		if tier := normalizeFishGenerateTier(req.Tier); tier != "" {
			args = append(args, "--tiers", tier)
		}
	}

	stdout, stderr, err := runFishGeneratorCommand(generatorDir, args)
	if err != nil {
		message := strings.TrimSpace(stderr)
		if message == "" {
			message = err.Error()
		}
		return nil, newConflictError("生成数据失败：" + message)
	}

	total := parseFishGeneratorTotal(stdout)
	message := strings.TrimSpace(stdout)
	if message == "" {
		message = "生成数据完成"
	}

	return &GenerateFishDataResponse{
		Mode:        mode,
		FishDirName: fishDirName,
		OutputDir:   filepath.ToSlash(filepath.Join("generated", fishDirName)),
		Total:       total,
		Message:     message,
	}, nil
}

func fishGeneratorDir() (string, error) {
	candidates := []string{
		filepath.Join("..", "senspace-web", "src", "components", "StarSky", "Desktop", "Plugins", "FishTank", "fish-generator"),
		filepath.Join("senspace-web", "src", "components", "StarSky", "Desktop", "Plugins", "FishTank", "fish-generator"),
	}
	for _, candidate := range candidates {
		absCandidate, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absCandidate); err == nil && info.IsDir() {
			return absCandidate, nil
		}
	}
	return "", newConflictError("FishTank 生成器目录不存在")
}

func runFishGeneratorCommand(dir string, args []string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "node", args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), fmt.Errorf("生成数据超时")
	}
	return stdout.String(), stderr.String(), err
}

func normalizeFishGenerateCount(count int) int {
	if count <= 0 {
		return 12
	}
	if count > 60000 {
		return 60000
	}
	return count
}

func normalizeFishGenerateTier(tier string) string {
	switch strings.TrimSpace(tier) {
	case "common", "rare", "epic", "legendary":
		return strings.TrimSpace(tier)
	default:
		return ""
	}
}

func parseFishGeneratorTotal(output string) int {
	start := strings.Index(output, "完成：")
	if start < 0 {
		return 0
	}
	rest := strings.TrimSpace(output[start+len("完成："):])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return 0
	}
	total, _ := strconv.Atoi(fields[0])
	return total
}
