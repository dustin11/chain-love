package task_service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"senspace/domain"
	"senspace/domain/task"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// 默认最大重试次数。
	defaultMaxRetry = 6
	// 默认优先级。
	defaultPriority = 100
	// 默认租约时长。
	defaultLeaseDuration = 30 * time.Second
)

// EnqueueRequest 入队参数。
type EnqueueRequest struct {
	// 任务类型。
	TaskType task.Type
	// 任务目标键。
	TaskKey string
	// 业务类型。
	BizType string
	// 业务主键。
	BizId int64
	// 去重键。
	DedupeKey string
	// 任务依据版本。
	SourceVersion string
	// 任务载荷。
	Payload any
	// 优先级，越小越高。
	Priority int
	// 最大重试次数。
	MaxRetry int
}

// Handler 任务执行函数。
type Handler func(*task.AsyncTask) error

var handlers = map[task.Type]Handler{}

// RegisterHandler 注册任务执行器。
func RegisterHandler(taskType task.Type, handler Handler) {
	if handler == nil {
		return
	}
	handlers[taskType] = handler
}

// Enqueue 幂等入队。
func Enqueue(req EnqueueRequest) (*task.AsyncTask, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeEnqueueRequest(req)
	if err != nil {
		return nil, err
	}

	var current task.AsyncTask
	err = tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("dedupe_key = ?", normalized.DedupeKey).
			First(&current).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			current = task.AsyncTask{
				Id:            generateTaskID(),
				TaskType:      normalized.TaskType,
				TaskKey:       normalized.TaskKey,
				BizType:       normalized.BizType,
				BizId:         normalized.BizId,
				Status:        task.StatusPending,
				Priority:      normalized.Priority,
				MaxRetry:      normalized.MaxRetry,
				PayloadJson:   normalized.PayloadJson,
				ResultJson:    "{}",
				DedupeKey:     normalized.DedupeKey,
				SourceVersion: normalized.SourceVersion,
			}
			return tx.Create(&current).Error
		}

		updates := map[string]any{
			"task_type":        normalized.TaskType,
			"task_key":         normalized.TaskKey,
			"biz_type":         normalized.BizType,
			"biz_id":           normalized.BizId,
			"payload_json":     normalized.PayloadJson,
			"priority":         normalized.Priority,
			"max_retry":        normalized.MaxRetry,
			"source_version":   normalized.SourceVersion,
			"next_retry_at":    nil,
			"lease_owner":      "",
			"lease_expires_at": nil,
		}
		if current.Status != task.StatusBuilding {
			updates["status"] = task.StatusPending
			updates["last_error"] = ""
			current.Status = task.StatusPending
			current.LastError = ""
		}
		current.TaskType = normalized.TaskType
		current.TaskKey = normalized.TaskKey
		current.BizType = normalized.BizType
		current.BizId = normalized.BizId
		current.PayloadJson = normalized.PayloadJson
		current.Priority = normalized.Priority
		current.MaxRetry = normalized.MaxRetry
		current.SourceVersion = normalized.SourceVersion
		current.LeaseOwner = ""
		current.LeaseExpiresAt = nil
		current.NextRetryAt = nil
		return tx.Model(&task.AsyncTask{}).
			Where("id = ?", current.Id).
			Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return &current, nil
}

// StartWorker 启动后台轮询 worker。
func StartWorker() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := RunPendingTasks(4); err != nil {
				continue
			}
		}
	}()
}

// RunPendingTasks 拉取并执行一批待执行任务。
func RunPendingTasks(limit int) error {
	if limit <= 0 {
		limit = 1
	}
	tasks, err := claimRunnableTasks(limit)
	if err != nil {
		return err
	}
	for i := range tasks {
		runTask(&tasks[i])
	}
	return nil
}

func runTask(item *task.AsyncTask) {
	handler, ok := handlers[item.TaskType]
	if !ok {
		_ = markTaskFailed(item.Id, fmt.Sprintf("任务类型未注册: %s", item.TaskType), false)
		return
	}
	if err := handler(item); err != nil {
		_ = markTaskFailed(item.Id, err.Error(), true)
		return
	}
	_ = markTaskSuccess(item.Id, "")
}

func claimRunnableTasks(limit int) ([]task.AsyncTask, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	leaseOwner := workerLeaseOwner()
	result := make([]task.AsyncTask, 0, limit)
	err = tx.Transaction(func(tx *gorm.DB) error {
		var items []task.AsyncTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status IN ?", []task.Status{task.StatusPending, task.StatusFailed}).
			Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
			Where("lease_expires_at IS NULL OR lease_expires_at <= ?", now).
			Order("priority ASC").
			Order("updated_at ASC").
			Limit(limit).
			Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			leaseExpiresAt := now.Add(defaultLeaseDuration)
			if err := tx.Model(&task.AsyncTask{}).
				Where("id = ?", item.Id).
				Updates(map[string]any{
					"status":           task.StatusBuilding,
					"lease_owner":      leaseOwner,
					"lease_expires_at": leaseExpiresAt,
					"started_at":       now,
				}).Error; err != nil {
				return err
			}
			item.Status = task.StatusBuilding
			item.LeaseOwner = leaseOwner
			item.StartedAt = &now
			item.LeaseExpiresAt = &leaseExpiresAt
			result = append(result, item)
		}
		return nil
	})
	return result, err
}

func markTaskSuccess(id int64, result string) error {
	tx, err := db()
	if err != nil {
		return err
	}
	now := time.Now()
	if strings.TrimSpace(result) == "" {
		result = "{}"
	}
	return tx.Model(&task.AsyncTask{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":           task.StatusReady,
			"result_json":      result,
			"last_error":       "",
			"lease_owner":      "",
			"lease_expires_at": nil,
			"finished_at":      now,
			"next_retry_at":    nil,
		}).Error
}

func markTaskFailed(id int64, message string, allowRetry bool) error {
	tx, err := db()
	if err != nil {
		return err
	}
	var item task.AsyncTask
	if err := tx.First(&item, "id = ?", id).Error; err != nil {
		return err
	}
	nextStatus := task.StatusDead
	nextRetryAt := any(nil)
	retryCount := item.RetryCount
	if allowRetry && retryCount < item.MaxRetry {
		retryCount += 1
		nextStatus = task.StatusFailed
		delay := time.Duration(retryCount*retryCount) * time.Second
		nextRetryAt = time.Now().Add(delay)
	}
	now := time.Now()
	return tx.Model(&task.AsyncTask{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":            nextStatus,
			"retry_count":       retryCount,
			"last_error":        truncateError(message),
			"lease_owner":       "",
			"lease_expires_at":  nil,
			"finished_at":       now,
			"next_retry_at":     nextRetryAt,
		}).Error
}

type normalizedEnqueueRequest struct {
	TaskType      task.Type
	TaskKey       string
	BizType       string
	BizId         int64
	DedupeKey     string
	SourceVersion string
	PayloadJson   string
	Priority      int
	MaxRetry      int
}

func normalizeEnqueueRequest(req EnqueueRequest) (normalizedEnqueueRequest, error) {
	taskType := task.Type(strings.TrimSpace(string(req.TaskType)))
	taskKey := strings.TrimSpace(req.TaskKey)
	dedupeKey := strings.TrimSpace(req.DedupeKey)
	if taskType == "" {
		return normalizedEnqueueRequest{}, fmt.Errorf("task type is empty")
	}
	if taskKey == "" {
		return normalizedEnqueueRequest{}, fmt.Errorf("task key is empty")
	}
	if dedupeKey == "" {
		dedupeKey = fmt.Sprintf("%s:%s", taskType, taskKey)
	}
	payloadJson := "{}"
	if req.Payload != nil {
		data, err := json.Marshal(req.Payload)
		if err != nil {
			return normalizedEnqueueRequest{}, err
		}
		payloadJson = string(data)
	}
	priority := req.Priority
	if priority <= 0 {
		priority = defaultPriority
	}
	maxRetry := req.MaxRetry
	if maxRetry <= 0 {
		maxRetry = defaultMaxRetry
	}
	return normalizedEnqueueRequest{
		TaskType:      taskType,
		TaskKey:       taskKey,
		BizType:       strings.TrimSpace(req.BizType),
		BizId:         req.BizId,
		DedupeKey:     dedupeKey,
		SourceVersion: strings.TrimSpace(req.SourceVersion),
		PayloadJson:   payloadJson,
		Priority:      priority,
		MaxRetry:      maxRetry,
	}, nil
}

func truncateError(message string) string {
	text := strings.TrimSpace(message)
	if len(text) <= 2000 {
		return text
	}
	return text[:2000]
}

func db() (*gorm.DB, error) {
	if domain.Db == nil {
		return nil, fmt.Errorf("task db not initialized")
	}
	return domain.Db, nil
}

func generateTaskID() int64 {
	return time.Now().UnixNano()
}

func workerLeaseOwner() string {
	return fmt.Sprintf("pid-%d", time.Now().UnixNano())
}
