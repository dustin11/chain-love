package factory_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"senspace/domain/factory"
	"senspace/domain/task"
	"senspace/pkg/setting"
	"senspace/service/task_service"

	"gorm.io/gorm"
)

const (
	// 工厂发布快照任务业务类型。
	staticTaskBizTypeFactoryRelease = "factory_release"
	// 工厂持有人快照任务业务类型。
	staticTaskBizTypeFactoryOwner = "factory_owner"
)

type staticTaskPayload struct {
	// 目标 ownerKey。
	OwnerKey string `json:"ownerKey,omitempty"`
	// 目标发布记录 ID。
	ReleaseId string `json:"releaseId,omitempty"`
	// 目标插件 ID。
	PluginId string `json:"pluginId,omitempty"`
	// 目标插件名称。
	PluginName string `json:"pluginName,omitempty"`
	// 目标铸造记录 ID。
	MintRecordId string `json:"mintRecordId,omitempty"`
	// 目标用户 ID。
	UserId string `json:"userId,omitempty"`
}

func init() {
	task_service.RegisterHandler(task.TypeMint, runFactoryOwnerAssetsSnapshotTask)
	task_service.RegisterHandler(task.TypePublish, runFactoryReleaseSnapshotTask)
	task_service.RegisterDeleteHandler(task.TypeMint, deleteFactoryOwnerAssetsSnapshotTask)
	task_service.RegisterDeleteHandler(task.TypePublish, deleteFactoryReleaseSnapshotTask)
	task_service.RegisterAccessChecker(task.TypeMint, checkFactoryOwnerAssetsSnapshotAccess)
	task_service.RegisterAccessChecker(task.TypePublish, checkFactoryReleaseSnapshotAccess)
}

// 入队持有人资产快照任务。
func enqueueFactoryOwnerAssetsSnapshot(ctx committedMintContext, pluginName string) error {
	ownerKey := strings.TrimSpace(ctx.OwnerKey)
	if ownerKey == "" {
		return nil
	}
	releaseId := strconv.FormatInt(ctx.ReleaseId, 10)
	taskKey := buildFactoryOwnerSnapshotTaskKey(ownerKey, ctx.ReleaseId)
	_, err := task_service.Enqueue(task_service.EnqueueRequest{
		TaskType:      task.TypeMint,
		TaskKey:       taskKey,
		BizType:       staticTaskBizTypeFactoryOwner,
		BizId:         ctx.ReleaseId,
		BizName:       strings.TrimSpace(pluginName),
		DedupeKey:     "factory-owner-assets:" + taskKey,
		SourceVersion: releaseId,
		Payload: staticTaskPayload{
			OwnerKey:     ownerKey,
			ReleaseId:    releaseId,
			PluginId:     strings.TrimSpace(ctx.PluginId),
			PluginName:   strings.TrimSpace(pluginName),
			MintRecordId: strconv.FormatInt(ctx.MintRecordId, 10),
			UserId:       strconv.FormatUint(ctx.UserId, 10),
		},
	})
	return err
}

// 同步重建持有人快照，并复用统一任务执行逻辑。
func rebuildOwnerFactorySnapshots(ownerKey string) error {
	ownerKey = strings.TrimSpace(ownerKey)
	if ownerKey == "" {
		return nil
	}
	payload := staticTaskPayload{OwnerKey: ownerKey}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return runFactoryOwnerAssetsSnapshotTask(&task.AsyncTask{
		TaskType:    task.TypeMint,
		TaskKey:     buildFactoryOwnerSnapshotTaskKey(ownerKey, 0),
		BizType:     staticTaskBizTypeFactoryOwner,
		BizName:     "",
		PayloadJson: string(data),
	})
}

// 入队发布静态快照任务。
func enqueueFactoryReleaseSnapshot(release factory.Release) error {
	_, err := task_service.Enqueue(task_service.EnqueueRequest{
		TaskType:      task.TypePublish,
		TaskKey:       buildFactoryReleaseSnapshotTaskKey(release.Id),
		BizType:       staticTaskBizTypeFactoryRelease,
		BizId:         release.Id,
		BizName:       strings.TrimSpace(release.Name),
		DedupeKey:     "factory-release-snapshot:" + strconv.FormatInt(release.Id, 10),
		SourceVersion: release.UpdatedAt.Format(timeLayoutSecond),
		Payload: staticTaskPayload{
			ReleaseId: strconv.FormatInt(release.Id, 10),
		},
	})
	return err
}

// 执行铸造任务，重建持有人资产快照。
func runFactoryOwnerAssetsSnapshotTask(item *task.AsyncTask) error {
	payload, err := parseStaticTaskPayload(item)
	if err != nil {
		return err
	}
	if strings.TrimSpace(payload.OwnerKey) == "" {
		return fmt.Errorf("ownerKey is empty")
	}
	return rebuildOwnerFactorySnapshotsNow(payload.OwnerKey)
}

// 执行发布任务，重建发布静态快照。
func runFactoryReleaseSnapshotTask(item *task.AsyncTask) error {
	payload, err := parseStaticTaskPayload(item)
	if err != nil {
		return err
	}
	releaseId, err := parseID(payload.ReleaseId, "发布记录ID")
	if err != nil {
		return err
	}
	tx, err := db()
	if err != nil {
		return err
	}
	var release factory.Release
	if err := tx.First(&release, "id = ?", releaseId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 发布记录不存在", task_service.ErrPermanentFailure)
		}
		return err
	}
	return rebuildReleaseStaticSnapshotNow(release)
}

// 删除发布任务时清理关联静态产物。
func deleteFactoryReleaseSnapshotTask(item *task.AsyncTask) error {
	releaseId, err := resolveStaticTaskReleaseID(item)
	if err != nil || releaseId == 0 {
		return nil
	}
	return purgeReleaseGeneratedArtifactsByID(releaseId)
}

// 删除铸造任务时清理关联快照或回滚铸造产物。
func deleteFactoryOwnerAssetsSnapshotTask(item *task.AsyncTask) error {
	payload, err := parseStaticTaskPayload(item)
	if err != nil {
		return err
	}
	mintRecordId, err := parseOptionalID(payload.MintRecordId)
	if err != nil {
		return err
	}
	releaseId, err := parseOptionalID(payload.ReleaseId)
	if err != nil {
		return err
	}
	userId, err := parseOptionalUintID(payload.UserId)
	if err != nil {
		return err
	}
	ownerKey := strings.TrimSpace(payload.OwnerKey)
	pluginId := strings.TrimSpace(payload.PluginId)

	if mintRecordId > 0 && releaseId > 0 && userId > 0 && ownerKey != "" && pluginId != "" {
		return rollbackCommittedMintArtifacts(committedMintContext{
			ReleaseId:    releaseId,
			PluginId:     pluginId,
			MintRecordId: mintRecordId,
			UserId:       userId,
			OwnerKey:     ownerKey,
		})
	}
	if ownerKey != "" {
		return removeOwnerFactorySnapshotArtifacts(ownerKey)
	}
	return nil
}

// 校验当前用户是否可操作发布任务。
func checkFactoryReleaseSnapshotAccess(item *task.AsyncTask, userId uint64) error {
	if item == nil || userId == 0 {
		return newForbiddenError("无权操作该任务")
	}
	releaseId, err := resolveStaticTaskReleaseID(item)
	if err != nil || releaseId == 0 {
		return newForbiddenError("无权操作该任务")
	}
	tx, err := db()
	if err != nil {
		return err
	}
	var release factory.Release
	if err := tx.Select("id", "author_id").First(&release, "id = ?", releaseId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("关联发布记录不存在")
		}
		return err
	}
	if release.AuthorId != userId {
		return newForbiddenError("无权操作该任务")
	}
	return nil
}

// 校验当前用户是否可操作铸造任务。
func checkFactoryOwnerAssetsSnapshotAccess(item *task.AsyncTask, userId uint64) error {
	if setting.IsDevLikeEnv() {
		return nil
	}
	if item == nil || userId == 0 {
		return newForbiddenError("无权操作该任务")
	}
	payload, err := parseStaticTaskPayload(item)
	if err != nil {
		return err
	}
	payloadUserId, err := parseOptionalUintID(payload.UserId)
	if err != nil {
		return err
	}
	if payloadUserId == 0 || payloadUserId != userId {
		return newForbiddenError("无权操作该任务")
	}
	return nil
}

// 解析任务载荷。
func parseStaticTaskPayload(item *task.AsyncTask) (staticTaskPayload, error) {
	var payload staticTaskPayload
	if strings.TrimSpace(item.PayloadJson) == "" {
		return payload, nil
	}
	if err := json.Unmarshal([]byte(item.PayloadJson), &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

// 构造发布任务键。
func buildFactoryReleaseSnapshotTaskKey(releaseId int64) string {
	return "release:" + strconv.FormatInt(releaseId, 10)
}

// 构造铸造任务键。
func buildFactoryOwnerSnapshotTaskKey(ownerKey string, releaseId int64) string {
	ownerKey = strings.TrimSpace(ownerKey)
	if releaseId <= 0 {
		return "owner:" + ownerKey
	}
	return "owner:" + ownerKey + ":release:" + strconv.FormatInt(releaseId, 10)
}

// 解析任务关联的发布记录 ID。
func resolveStaticTaskReleaseID(item *task.AsyncTask) (int64, error) {
	if item == nil {
		return 0, nil
	}
	if item.BizId > 0 {
		return item.BizId, nil
	}
	payload, err := parseStaticTaskPayload(item)
	if err != nil {
		return 0, err
	}
	return parseOptionalID(payload.ReleaseId)
}

// 解析可选 int64 ID。
func parseOptionalID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	return parseID(raw, "ID")
}

// 解析可选 uint64 ID。
func parseOptionalUintID(raw string) (uint64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseUint(raw, 10, 64)
}
