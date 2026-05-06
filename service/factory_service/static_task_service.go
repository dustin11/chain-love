package factory_service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"senspace/domain/factory"
	"senspace/domain/task"
	"senspace/service/task_service"
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
}

func init() {
	task_service.RegisterHandler(task.TypeFactoryOwnerAssetsSnapshot, runFactoryOwnerAssetsSnapshotTask)
	task_service.RegisterHandler(task.TypeFactoryReleaseSnapshot, runFactoryReleaseSnapshotTask)
}

// 入队持有人资产快照任务。
func enqueueFactoryOwnerAssetsSnapshot(ownerKey string) error {
	ownerKey = strings.TrimSpace(ownerKey)
	if ownerKey == "" {
		return nil
	}
	_, err := task_service.Enqueue(task_service.EnqueueRequest{
		TaskType:      task.TypeFactoryOwnerAssetsSnapshot,
		TaskKey:       "owner:" + ownerKey,
		BizType:       staticTaskBizTypeFactoryOwner,
		DedupeKey:     "factory-owner-assets:" + ownerKey,
		SourceVersion: ownerKey,
		Payload: staticTaskPayload{
			OwnerKey: ownerKey,
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
		TaskType:    task.TypeFactoryOwnerAssetsSnapshot,
		TaskKey:     "owner:" + ownerKey,
		BizType:     staticTaskBizTypeFactoryOwner,
		PayloadJson: string(data),
	})
}

// 入队发布静态快照任务。
func enqueueFactoryReleaseSnapshot(release factory.Release) error {
	_, err := task_service.Enqueue(task_service.EnqueueRequest{
		TaskType:      task.TypeFactoryReleaseSnapshot,
		TaskKey:       buildFactoryReleaseSnapshotTaskKey(release.Id),
		BizType:       staticTaskBizTypeFactoryRelease,
		BizId:         release.Id,
		DedupeKey:     "factory-release-snapshot:" + strconv.FormatInt(release.Id, 10),
		SourceVersion: release.UpdatedAt.Format(timeLayoutSecond),
		Payload: staticTaskPayload{
			ReleaseId: strconv.FormatInt(release.Id, 10),
		},
	})
	return err
}

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
		return err
	}
	return rebuildReleaseStaticSnapshotNow(release)
}

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

func buildFactoryReleaseSnapshotTaskKey(releaseId int64) string {
	return "release:" + strconv.FormatInt(releaseId, 10)
}
