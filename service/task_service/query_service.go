package task_service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"senspace/domain/task"
	"senspace/pkg/setting"

	"gorm.io/gorm"
)

// ListTasks 查询任务列表。
func ListTasks(query Query) (*TaskListResult, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	dbq := tx.Model(&task.AsyncTask{})
	if value := strings.TrimSpace(query.TaskType); value != "" {
		dbq = dbq.Where("task_type = ?", value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		dbq = dbq.Where("status = ?", value)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		dbq = dbq.Where(
			"task_key LIKE ? OR biz_type LIKE ? OR biz_name LIKE ? OR last_error LIKE ?",
			like,
			like,
			like,
			like,
		)
	}

	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []task.AsyncTask
	if err := dbq.
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, err
	}

	result := &TaskListResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    make([]TaskView, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, mapTaskView(item))
	}
	return result, nil
}

// RetryTaskNow 手动重试任务。
func RetryTaskNow(idRaw string, userId uint64) (*TaskView, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(idRaw), 10, 64)
	if err != nil || id <= 0 {
		return nil, errors.New("任务ID无效")
	}
	tx, err := db()
	if err != nil {
		return nil, err
	}
	var item task.AsyncTask
	if err := tx.First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务不存在")
		}
		return nil, err
	}
	if err := ensureTaskAccess(&item, userId); err != nil {
		return nil, err
	}
	if err := tx.Model(&task.AsyncTask{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":           task.StatusPending,
			"last_error":       "",
			"lease_owner":      "",
			"lease_expires_at": nil,
			"next_retry_at":    nil,
			"finished_at":      nil,
		}).Error; err != nil {
		return nil, err
	}
	item.Status = task.StatusPending
	item.LastError = ""
	item.LeaseOwner = ""
	item.LeaseExpiresAt = nil
	item.NextRetryAt = nil
	item.FinishedAt = nil
	return toTaskViewPtr(item), nil
}

// DeleteTask 删除任务及其关联清理数据。
func DeleteTask(idRaw string, userId uint64) error {
	id, err := strconv.ParseInt(strings.TrimSpace(idRaw), 10, 64)
	if err != nil || id <= 0 {
		return errors.New("任务ID无效")
	}
	tx, err := db()
	if err != nil {
		return err
	}

	var item task.AsyncTask
	if err := tx.First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("任务不存在")
		}
		return err
	}
	if err := ensureTaskAccess(&item, userId); err != nil {
		if !isIgnorableDeleteDependencyError(err) {
			return err
		}
	}
	if !setting.IsDevLikeEnv() {
		if item.Status == task.StatusBuilding {
			return errors.New("执行中的任务不能删除")
		}
		if item.Status != task.StatusFailed && item.Status != task.StatusDead {
			return errors.New("仅支持删除失败任务")
		}
	}

	if handler := deleteHandlers[item.TaskType]; handler != nil {
		if err := handler(&item); err != nil {
			if !isIgnorableDeleteDependencyError(err) {
				return err
			}
		}
	}
	return tx.Delete(&task.AsyncTask{}, "id = ?", id).Error
}

func ensureTaskAccess(item *task.AsyncTask, userId uint64) error {
	if item == nil {
		return errors.New("任务不存在")
	}
	if userId == 0 {
		return errors.New("无权操作该任务")
	}
	if checker := accessCheckers[item.TaskType]; checker != nil {
		return checker(item, userId)
	}
	return nil
}

func isIgnorableDeleteDependencyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	return strings.Contains(err.Error(), "关联发布记录不存在")
}

func mapTaskView(item task.AsyncTask) TaskView {
	return TaskView{
		Id:            strconv.FormatInt(item.Id, 10),
		TaskType:      item.TaskType,
		TaskKey:       item.TaskKey,
		BizType:       item.BizType,
		BizId:         formatTaskBizId(item.BizId),
		BizName:       strings.TrimSpace(item.BizName),
		Status:        item.Status,
		Priority:      item.Priority,
		RetryCount:    item.RetryCount,
		MaxRetry:      item.MaxRetry,
		LastError:     item.LastError,
		SourceVersion: item.SourceVersion,
		LeaseOwner:    item.LeaseOwner,
		StartedAt:     formatTaskTime(item.StartedAt),
		FinishedAt:    formatTaskTime(item.FinishedAt),
		NextRetryAt:   formatTaskTime(item.NextRetryAt),
		CreatedAt:     item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:     item.UpdatedAt.Format(time.RFC3339Nano),
	}
}

// toTaskViewPtr 转成指针视图，便于接口直接返回。
func toTaskViewPtr(item task.AsyncTask) *TaskView {
	view := mapTaskView(item)
	return &view
}

func formatTaskBizId(id int64) string {
	if id <= 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

func formatTaskTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}
