package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"senspace/domain"
	"senspace/domain/d_util"
	factorydomain "senspace/domain/factory"
	"senspace/middleware"
	"senspace/pkg/app/security"
	"senspace/pkg/setting"
	"senspace/pkg/setting/consts"
	"senspace/pkg/util"
	"senspace/routers"
	"senspace/service/factory_service"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

const (
	factoryPublishPluginID = "990000000000001001"
	factoryUpgradePluginID = "990000000000001002"
)

type testAPIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type factoryTestEnv struct {
	t                 *testing.T
	router            *gin.Engine
	db                *gorm.DB
	pluginRoot        string
	pluginSourceRoot  string
	pluginRuntimeRoot string
	user              security.JwtUser
}

type fakePluginBuildExecutor struct{}

func (fakePluginBuildExecutor) Build(ctx context.Context, req factory_service.PluginBuildRequest) (*factory_service.PluginBuildResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err := os.MkdirAll(filepath.Join(req.OutputDir, "runtime"), 0o755); err != nil {
		return nil, err
	}
	entryContent := []byte("export const plugin = 'built';\n")
	if err := os.WriteFile(filepath.Join(req.OutputDir, "runtime", "index.js"), entryContent, 0o644); err != nil {
		return nil, err
	}

	manifest := map[string]any{
		"pluginId":   req.PluginId,
		"version":    req.Version,
		"releaseId":  strconv.FormatInt(req.ReleaseId, 10),
		"bundleHash": "sha256:test-bundle-" + strconv.FormatInt(req.ReleaseId, 10),
		"integrity":  "sha384:test-integrity-" + strconv.FormatInt(req.ReleaseId, 10),
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(req.OutputDir, "runtime-manifest.json"), manifestBytes, 0o644); err != nil {
		return nil, err
	}

	return &factory_service.PluginBuildResult{
		BundleHash: manifest["bundleHash"].(string),
		Integrity:  manifest["integrity"].(string),
		BuiltAt:    time.Now(),
	}, nil
}

// 发布与管理流程。
func TestFactoryPublishAndManageRelease(t *testing.T) {
	env := setupFactoryTestEnv(t)

	pluginId := env.preparePlugin(factoryPublishPluginID, "Glass Ball Plugin", "玻璃球示例插件", "1.0.0", map[string]string{
		"src/index.ts":  "export const plugin = 'v1';\n",
		"src/helper.ts": "export const helper = true;\n",
	})

	publishPayload := factory_service.PublishRequest{
		PluginId: pluginId,
		Manifest: factory_service.PluginManifest{
			Name:        "Glass Ball Plugin",
			Version:     "1.0.0",
			Entry:       "src/index.ts",
			Description: "玻璃球示例插件",
		},
		Release: factory_service.ReleasePayload{
			MutableMarketMetadata: factory_service.MutableMarketMetadata{
				Summary:  "一个适合桌面场景的玻璃球互动插件",
				Category: "decor",
				Tags:     []string{"桌面", "装饰", "互动"},
				CoverUrl: "https://cdn.example.com/factory/glass-ball-cover.png",
			},
			TotalSupply:   1000,
			MintPer:       5,
			MintPrice:     "0.05",
			UpgradePolicy: factorydomain.ReleaseUpgradePolicyPaid,
			UpgradePrice:  "0.01",
		},
	}

	resp := env.doJSON(http.MethodPost, "/api/v1/factory/publish", publishPayload)
	require.Equal(t, http.StatusOK, resp.Code)

	var publishResp testAPIResponse
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &publishResp))
	require.Equal(t, 0, publishResp.Code)

	var created factory_service.PublishRecord
	require.NoError(t, json.Unmarshal(publishResp.Data, &created))
	require.NotEmpty(t, created.Id)
	require.Equal(t, pluginId, created.PluginId)
	require.Equal(t, factorydomain.ReleaseStatusPublished, created.Status)
	require.True(t, created.CurrentRelease)
	require.Equal(t, factorydomain.BuildStatusReady, created.BuildStatus)
	require.Equal(t, "10001", created.Author.Id)
	require.Equal(t, "factory-tester", created.Author.Name)
	require.Equal(t, "0.05", created.MintPrice)
	require.Equal(t, "0.01", created.UpgradePrice)
	require.True(t, strings.HasPrefix(created.SourceHash, "sha256:"))
	require.True(t, strings.HasPrefix(created.BundleHash, "sha256:"))
	require.True(t, strings.HasPrefix(created.Integrity, "sha384:"))
	require.NotEmpty(t, created.BuiltAt)
	require.FileExists(t, filepath.Join(env.pluginSourceRoot, pluginId, created.Id, "manifest.snapshot.json"))
	require.FileExists(t, filepath.Join(env.pluginRuntimeRoot, pluginId, "v1.0.0-"+created.Id, "runtime", "index.js"))

	dupResp := env.doJSON(http.MethodPost, "/api/v1/factory/publish", publishPayload)
	require.Equal(t, http.StatusBadRequest, dupResp.Code)
	require.NoError(t, json.Unmarshal(dupResp.Body.Bytes(), &publishResp))
	require.Equal(t, 200409, publishResp.Code)

	listResp := env.doJSON(http.MethodGet, "/api/v1/factory/releases/my?currentOnly=true", nil)
	require.Equal(t, http.StatusOK, listResp.Code)
	require.NoError(t, json.Unmarshal(listResp.Body.Bytes(), &publishResp))
	var myReleases []factory_service.PublishRecord
	require.NoError(t, json.Unmarshal(publishResp.Data, &myReleases))
	require.Len(t, myReleases, 1)
	require.Equal(t, created.Id, myReleases[0].Id)
	require.Equal(t, "10001", myReleases[0].Author.Id)

	marketResp := env.doJSON(http.MethodGet, "/api/v1/factory/market", nil)
	require.Equal(t, http.StatusOK, marketResp.Code)
	require.NoError(t, json.Unmarshal(marketResp.Body.Bytes(), &publishResp))
	var marketList []factory_service.PublishRecord
	require.NoError(t, json.Unmarshal(publishResp.Data, &marketList))
	require.NotEmpty(t, marketList)
	matched := false
	for _, item := range marketList {
		if item.Id == created.Id {
			require.Equal(t, "10001", item.Author.Id)
			matched = true
			break
		}
	}
	require.True(t, matched, "market list should contain created release")

	detailResp := env.doJSON(http.MethodGet, "/api/v1/factory/releases/"+created.Id, nil)
	require.Equal(t, http.StatusOK, detailResp.Code)
	require.NoError(t, json.Unmarshal(detailResp.Body.Bytes(), &publishResp))
	var detail factory_service.PublishDetail
	require.NoError(t, json.Unmarshal(publishResp.Data, &detail))
	require.Equal(t, "1.0.0", detail.ManifestSnapshot.Version)
	require.Empty(t, detail.PriceHistory)

	updateMarket := factory_service.UpdateReleaseRequest{
		Market: factory_service.MutableMarketMetadata{
			Summary:  "更新后的市场摘要",
			Category: "decor",
			Tags:     []string{"桌面", "收藏"},
			CoverUrl: "https://cdn.example.com/new-cover.png",
		},
	}
	updateMarketResp := env.doJSON(http.MethodPatch, "/api/v1/factory/releases/"+created.Id, updateMarket)
	require.Equal(t, http.StatusOK, updateMarketResp.Code)

	updatePrice := factory_service.UpdateReleasePriceRequest{
		MintPrice: "0.08",
		Reason:    "节日活动调价",
	}
	updatePriceResp := env.doJSON(http.MethodPatch, "/api/v1/factory/releases/"+created.Id+"/price", updatePrice)
	require.Equal(t, http.StatusOK, updatePriceResp.Code)

	detailResp = env.doJSON(http.MethodGet, "/api/v1/factory/releases/"+created.Id, nil)
	require.Equal(t, http.StatusOK, detailResp.Code)
	require.NoError(t, json.Unmarshal(detailResp.Body.Bytes(), &publishResp))
	require.NoError(t, json.Unmarshal(publishResp.Data, &detail))
	require.Equal(t, "更新后的市场摘要", detail.Summary)
	require.Equal(t, []string{"桌面", "收藏"}, detail.Tags)
	require.Equal(t, "0.08", detail.MintPrice)
	require.Len(t, detail.PriceHistory, 1)
	require.Equal(t, "0.05", detail.PriceHistory[0].PreviousMintPrice)
	require.Equal(t, "0.08", detail.PriceHistory[0].NextMintPrice)

	pausedResp := env.doJSON(http.MethodPatch, "/api/v1/factory/releases/"+created.Id+"/status", factory_service.UpdateReleaseStatusRequest{
		Status: factorydomain.ReleaseStatusPaused,
		Reason: "临时维护",
	})
	require.Equal(t, http.StatusOK, pausedResp.Code)

	closedResp := env.doJSON(http.MethodPatch, "/api/v1/factory/releases/"+created.Id+"/status", factory_service.UpdateReleaseStatusRequest{
		Status: factorydomain.ReleaseStatusClosed,
		Reason: "停止销售",
	})
	require.Equal(t, http.StatusOK, closedResp.Code)

	reopenResp := env.doJSON(http.MethodPatch, "/api/v1/factory/releases/"+created.Id+"/status", factory_service.UpdateReleaseStatusRequest{
		Status: factorydomain.ReleaseStatusPublished,
		Reason: "尝试重新上架",
	})
	require.Equal(t, http.StatusBadRequest, reopenResp.Code)
	require.NoError(t, json.Unmarshal(reopenResp.Body.Bytes(), &publishResp))
	require.Equal(t, 200409, publishResp.Code)
}

// 铸造与升级流程。
func TestFactoryOwnershipUpgradeFlow(t *testing.T) {
	env := setupFactoryTestEnv(t)

	pluginId := env.preparePlugin(factoryUpgradePluginID, "Upgrade Plugin", "升级插件", "1.0.0", map[string]string{
		"src/index.ts": "export const plugin = 'v1';\n",
	})

	releaseV1 := env.publish(factory_service.PublishRequest{
		PluginId: pluginId,
		Manifest: factory_service.PluginManifest{
			Name:        "Upgrade Plugin",
			Version:     "1.0.0",
			Entry:       "src/index.ts",
			Description: "升级插件",
		},
		Release: factory_service.ReleasePayload{
			MutableMarketMetadata: factory_service.MutableMarketMetadata{
				Summary:  "v1 发布",
				Category: "tool",
				Tags:     []string{"升级"},
			},
			TotalSupply:   10,
			MintPer:       1,
			MintPrice:     "0.05",
			UpgradePolicy: factorydomain.ReleaseUpgradePolicyNone,
		},
	})

	_, err := factory_service.RecordMint(factory_service.RecordMintRequest{
		ReleaseId:     releaseV1.Id,
		UserId:        env.user.Id,
		WalletAddress: env.user.Addr,
		Quantity:      2,
		TotalPaid:     "0.10",
	})
	require.Error(t, err)

	mintRecord, err := factory_service.RecordMint(factory_service.RecordMintRequest{
		ReleaseId:     releaseV1.Id,
		UserId:        env.user.Id,
		WalletAddress: env.user.Addr,
		Quantity:      1,
		TotalPaid:     "0.05",
	})
	require.NoError(t, err)
	require.Equal(t, pluginId, mintRecord.PluginId)

	ownerships, err := factory_service.ListMyOwnerships(env.user.Id)
	require.NoError(t, err)
	require.Len(t, ownerships, 1)
	require.Equal(t, factorydomain.OwnershipUpgradeStateUpToDate, ownerships[0].UpgradeState)

	env.writePluginVersion(pluginId, "Upgrade Plugin", "升级插件 1.1", "1.1.0", map[string]string{
		"src/index.ts": "export const plugin = 'v1.1';\n",
	})
	releaseV11 := env.publish(factory_service.PublishRequest{
		PluginId: pluginId,
		Manifest: factory_service.PluginManifest{
			Name:        "Upgrade Plugin",
			Version:     "1.1.0",
			Entry:       "src/index.ts",
			Description: "升级插件 1.1",
		},
		Release: factory_service.ReleasePayload{
			MutableMarketMetadata: factory_service.MutableMarketMetadata{
				Summary:  "v1.1 发布",
				Category: "tool",
				Tags:     []string{"升级"},
			},
			TotalSupply:   10,
			MintPer:       1,
			MintPrice:     "0.06",
			UpgradePolicy: factorydomain.ReleaseUpgradePolicyMajorPaid,
			UpgradePrice:  "0.20",
		},
	})

	ownershipResp := env.doJSON(http.MethodGet, "/api/v1/factory/ownership/my", nil)
	require.Equal(t, http.StatusOK, ownershipResp.Code)
	var apiResp testAPIResponse
	require.NoError(t, json.Unmarshal(ownershipResp.Body.Bytes(), &apiResp))
	var ownershipViews []factory_service.UserPluginOwnershipView
	require.NoError(t, json.Unmarshal(apiResp.Data, &ownershipViews))
	require.Len(t, ownershipViews, 1)
	require.Equal(t, factorydomain.OwnershipUpgradeStateUpgradable, ownershipViews[0].UpgradeState)
	require.Equal(t, releaseV11.Id, ownershipViews[0].LatestAvailableReleaseId)
	require.Equal(t, "0", ownershipViews[0].UpgradePrice)

	upgradeResp := env.doJSON(http.MethodPost, "/api/v1/factory/ownership/"+ownershipViews[0].Id+"/upgrade", map[string]string{
		"toReleaseId": releaseV11.Id,
	})
	require.Equal(t, http.StatusOK, upgradeResp.Code)
	require.NoError(t, json.Unmarshal(upgradeResp.Body.Bytes(), &apiResp))
	var upgradeRecord factory_service.UpgradeRecord
	require.NoError(t, json.Unmarshal(apiResp.Data, &upgradeRecord))
	require.Equal(t, factorydomain.UpgradeTypeFree, upgradeRecord.UpgradeType)
	require.Equal(t, "0", upgradeRecord.PaidAmount)

	env.writePluginVersion(pluginId, "Upgrade Plugin", "升级插件 2.0", "2.0.0", map[string]string{
		"src/index.ts": "export const plugin = 'v2';\n",
	})
	releaseV2 := env.publish(factory_service.PublishRequest{
		PluginId: pluginId,
		Manifest: factory_service.PluginManifest{
			Name:        "Upgrade Plugin",
			Version:     "2.0.0",
			Entry:       "src/index.ts",
			Description: "升级插件 2.0",
		},
		Release: factory_service.ReleasePayload{
			MutableMarketMetadata: factory_service.MutableMarketMetadata{
				Summary:  "v2 发布",
				Category: "tool",
				Tags:     []string{"升级"},
			},
			TotalSupply:   10,
			MintPer:       1,
			MintPrice:     "0.07",
			UpgradePolicy: factorydomain.ReleaseUpgradePolicyMajorPaid,
			UpgradePrice:  "0.50",
		},
	})

	ownershipResp = env.doJSON(http.MethodGet, "/api/v1/factory/ownership/my", nil)
	require.Equal(t, http.StatusOK, ownershipResp.Code)
	require.NoError(t, json.Unmarshal(ownershipResp.Body.Bytes(), &apiResp))
	require.NoError(t, json.Unmarshal(apiResp.Data, &ownershipViews))
	require.Len(t, ownershipViews, 1)
	require.Equal(t, factorydomain.OwnershipUpgradeStateUpgradeRequired, ownershipViews[0].UpgradeState)
	require.Equal(t, releaseV2.Id, ownershipViews[0].LatestAvailableReleaseId)
	require.Equal(t, "0.5", ownershipViews[0].UpgradePrice)

	upgradeResp = env.doJSON(http.MethodPost, "/api/v1/factory/ownership/"+ownershipViews[0].Id+"/upgrade", map[string]string{
		"toReleaseId": releaseV2.Id,
	})
	require.Equal(t, http.StatusOK, upgradeResp.Code)
	require.NoError(t, json.Unmarshal(upgradeResp.Body.Bytes(), &apiResp))
	require.NoError(t, json.Unmarshal(apiResp.Data, &upgradeRecord))
	require.Equal(t, factorydomain.UpgradeTypePaid, upgradeRecord.UpgradeType)
	require.Equal(t, "0.5", upgradeRecord.PaidAmount)

	ownershipResp = env.doJSON(http.MethodGet, "/api/v1/factory/ownership/my", nil)
	require.Equal(t, http.StatusOK, ownershipResp.Code)
	require.NoError(t, json.Unmarshal(ownershipResp.Body.Bytes(), &apiResp))
	require.NoError(t, json.Unmarshal(apiResp.Data, &ownershipViews))
	require.Len(t, ownershipViews, 1)
	require.Equal(t, factorydomain.OwnershipUpgradeStateUpToDate, ownershipViews[0].UpgradeState)
	require.Equal(t, releaseV2.Id, ownershipViews[0].EffectiveReleaseId)
}

// dev 配置与测试数据。
func setupFactoryTestEnv(t *testing.T) *factoryTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	util.Setup()

	originalEnv := os.Getenv("SPIDER_ENV")
	originalConfig := *setting.Config
	originalDB := domain.Db

	require.NoError(t, os.Setenv("SPIDER_ENV", "dev"))
	loadFactoryDevConfig(t)

	pluginRoot := t.TempDir()
	pluginSourceRoot := t.TempDir()
	pluginRuntimeRoot := t.TempDir()

	// 数据库使用 dev 配置，插件和构建目录改为临时路径。
	setting.Config.App.FilePath.Plugin = pluginRoot
	setting.Config.App.PluginSourceRoot = pluginSourceRoot
	setting.Config.App.PluginRuntimeRoot = pluginRuntimeRoot
	setting.Config.App.PluginBuilderImage = "senspace/plugin-builder:test"

	domain.Setup()
	require.NotNil(t, domain.Db)
	d_util.InitTable(domain.Db)
	executeSQLFile(t, domain.Db, filepath.Join(repoRoot(t), "asset", "database", "factory_test_seed.sql"))

	router := gin.New()
	router.Use(middleware.ErrHandler())
	routers.SetupApiV1Router(router)

	resetBuildExecutor := factory_service.SetPluginBuildExecutorForTest(fakePluginBuildExecutor{})

	env := &factoryTestEnv{
		t:                 t,
		router:            router,
		db:                domain.Db,
		pluginRoot:        pluginRoot,
		pluginSourceRoot:  pluginSourceRoot,
		pluginRuntimeRoot: pluginRuntimeRoot,
		user: security.JwtUser{
			Id:       10001,
			Addr:     "0x1234567890",
			Nickname: "factory-tester",
		},
	}

	t.Cleanup(func() {
		if sqlDB, err := env.db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		domain.Db = originalDB
		*setting.Config = originalConfig
		resetBuildExecutor()
		if originalEnv == "" {
			_ = os.Unsetenv("SPIDER_ENV")
		} else {
			_ = os.Setenv("SPIDER_ENV", originalEnv)
		}
	})

	return env
}

// 读取 dev 配置。
func loadFactoryDevConfig(t *testing.T) {
	t.Helper()

	confPath := filepath.Join(repoRoot(t), "asset", "conf", "dev.yml")
	conf, err := os.ReadFile(confPath)
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(conf, setting.Config))

	resolved := strings.ReplaceAll(string(conf), "${s90}", setting.Config.S90)
	require.NoError(t, yaml.Unmarshal([]byte(resolved), setting.Config))
}

// 执行 SQL 文件。
func executeSQLFile(t *testing.T, db *gorm.DB, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	for _, stmt := range splitSQLStatements(string(content)) {
		require.NoError(t, db.Exec(stmt).Error)
	}
}

// 拆分 SQL 语句。
func splitSQLStatements(content string) []string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		filtered = append(filtered, line)
	}

	rawStatements := strings.Split(strings.Join(filtered, "\n"), ";")
	statements := make([]string, 0, len(rawStatements))
	for _, stmt := range rawStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		statements = append(statements, stmt)
	}
	return statements
}

// 仓库根目录。
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Clean(filepath.Join(wd, ".."))
}

// 写入插件源码目录。
func (env *factoryTestEnv) preparePlugin(pluginID string, name string, description string, version string, files map[string]string) string {
	env.t.Helper()
	env.writePluginVersion(pluginID, name, description, version, files)
	return pluginID
}

// 写入指定版本源码。
func (env *factoryTestEnv) writePluginVersion(pluginID string, name string, description string, version string, files map[string]string) {
	env.t.Helper()
	versionRoot := filepath.Join(env.pluginRoot, pluginID, version)
	require.NoError(env.t, os.MkdirAll(versionRoot, 0o755))

	manifest := map[string]any{
		"name":        name,
		"version":     version,
		"entry":       "src/index.ts",
		"description": description,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(env.t, err)
	require.NoError(env.t, os.WriteFile(filepath.Join(versionRoot, "manifest.json"), manifestBytes, 0o644))

	for rel, content := range files {
		fullPath := filepath.Join(versionRoot, filepath.FromSlash(rel))
		require.NoError(env.t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(env.t, os.WriteFile(fullPath, []byte(content), 0o644))
	}
}

// 通过 HTTP 创建发布。
func (env *factoryTestEnv) publish(payload factory_service.PublishRequest) factory_service.PublishRecord {
	env.t.Helper()
	resp := env.doJSON(http.MethodPost, "/api/v1/factory/publish", payload)
	require.Equal(env.t, http.StatusOK, resp.Code)

	var apiResp testAPIResponse
	require.NoError(env.t, json.Unmarshal(resp.Body.Bytes(), &apiResp))
	require.Equal(env.t, 0, apiResp.Code)

	var record factory_service.PublishRecord
	require.NoError(env.t, json.Unmarshal(apiResp.Data, &record))
	return record
}

// 发送 JSON 请求。
func (env *factoryTestEnv) doJSON(method string, path string, body any) *httptest.ResponseRecorder {
	env.t.Helper()

	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		require.NoError(env.t, err)
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(requestBody))
	require.NoError(env.t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(env.authCookie())

	recorder := httptest.NewRecorder()
	env.router.ServeHTTP(recorder, req)
	return recorder
}

// 生成测试登录 Cookie。
func (env *factoryTestEnv) authCookie() *http.Cookie {
	token, err := security.GenerateToken(env.user)
	require.NoError(env.t, err)
	return &http.Cookie{
		Name:  consts.ACCESS_TOKEN,
		Value: token,
		Path:  "/",
	}
}
