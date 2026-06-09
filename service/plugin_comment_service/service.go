package plugin_comment_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"senspace/domain"
	"senspace/domain/active"
	"senspace/domain/ds"
	"senspace/domain/ds/enum"
	"senspace/domain/sys"
	"senspace/pkg/app/security"

	"gorm.io/gorm"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

func db() (*gorm.DB, error) {
	if domain.Db == nil {
		return nil, fmt.Errorf("plugin comment db not initialized")
	}
	return domain.Db, nil
}

// ListComments 查询插件评论，默认返回最新 50 条。
func ListComments(req ListRequest, user *security.JwtUser) (*ListResult, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}
	instanceId := strings.TrimSpace(req.InstanceId)
	if instanceId == "" {
		instanceId = anchorString(req.Anchor, "instanceId")
	}
	if instanceId == "" {
		return nil, errors.New("instanceId不能为空")
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)

	query := tx.Model(&ds.PluginComment{}).Where("instance_id = ?", instanceId)
	if strings.TrimSpace(req.Mode) == "current" {
		if itemID := anchorItemID(req.Anchor); itemID != "" {
			query = query.Where("anchor_item_id = ?", itemID)
		} else if index, ok := anchorIndex(req.Anchor); ok {
			query = query.Where("anchor_index = ?", index)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var comments []ds.PluginComment
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		return nil, err
	}

	items, err := buildCommentViews(tx, comments, user)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}

// CreateComment 创建评论或回复。
func CreateComment(req CreateRequest, user *security.JwtUser) (*CommentView, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}
	if user == nil || user.Id == 0 {
		return nil, errors.New("请先登录")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("评论内容不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, errors.New("评论内容过长")
	}
	instanceId := anchorString(req.Anchor, "instanceId")
	if instanceId == "" {
		return nil, errors.New("评论锚点缺少 instanceId")
	}
	anchorJson, err := marshalAnchor(req.Anchor)
	if err != nil {
		return nil, err
	}
	anchorIdx, hasIndex := anchorIndex(req.Anchor)
	anchorItemID := anchorItemID(req.Anchor)
	comment := ds.PluginComment{
		InstanceId:      instanceId,
		AnchorJson:      anchorJson,
		Content:         content,
		ReplyToNickname: strings.TrimSpace(req.ReplyToNickname),
		CreatInfo: domain.CreatInfo{
			CreatedBy: user.Id,
		},
	}
	if hasIndex {
		comment.AnchorIndex = &anchorIdx
	}
	if anchorItemID != "" {
		comment.AnchorItemId = &anchorItemID
	}

	if strings.TrimSpace(req.ParentId) != "" {
		parentId, err := parseUintID(req.ParentId)
		if err != nil {
			return nil, err
		}
		var parent ds.PluginComment
		if err := tx.First(&parent, "id = ? AND instance_id = ?", parentId, instanceId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("回复目标不存在")
			}
			return nil, err
		}
		comment.ParentId = &parent.Id
		rootId := parent.Id
		if parent.RootId != nil {
			rootId = *parent.RootId
		}
		comment.RootId = &rootId
		comment.Level = parent.Level + 1
	}

	if err := tx.Transaction(func(trx *gorm.DB) error {
		if err := trx.Create(&comment).Error; err != nil {
			return err
		}
		if comment.RootId != nil {
			return trx.Model(&ds.PluginComment{}).
				Where("id = ?", *comment.RootId).
				UpdateColumn("reply_cnt", gorm.Expr("reply_cnt + 1")).Error
		}
		return nil
	}); err != nil {
		return nil, err
	}

	views, err := buildCommentViews(tx, []ds.PluginComment{comment}, user)
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, errors.New("评论创建失败")
	}
	return &views[0], nil
}

// CleanupComments 级联删除插件评论及其点赞数据。
func CleanupComments(req CleanupRequest, user *security.JwtUser) (*CleanupResult, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}
	if user == nil || user.Id == 0 {
		return nil, errors.New("请先登录")
	}
	instanceID := strings.TrimSpace(req.InstanceId)
	if instanceID == "" {
		return nil, errors.New("instanceId不能为空")
	}
	if req.DeleteAll {
		return cleanupAllInstanceComments(tx, instanceID)
	}
	if len(req.Items) == 0 {
		return &CleanupResult{}, nil
	}

	commentIDs := make([]uint64, 0)
	for _, item := range req.Items {
		commentID := strings.TrimSpace(item.CommentId)
		if commentID != "" {
			ids, queryErr := listCommentSubtreeIDsByCommentID(
				tx,
				instanceID,
				commentID,
				user.Id,
			)
			if queryErr != nil {
				return nil, queryErr
			}
			commentIDs = append(commentIDs, ids...)
			continue
		}

		itemID := strings.TrimSpace(item.ItemId)
		if itemID != "" {
			ids, queryErr := listCommentIDsByItemID(tx, instanceID, itemID)
			if queryErr != nil {
				return nil, queryErr
			}
			commentIDs = append(commentIDs, ids...)
			continue
		}

		if item.Index == nil {
			continue
		}
		ids, queryErr := listCommentIDsByIndex(tx, instanceID, *item.Index)
		if queryErr != nil {
			return nil, queryErr
		}
		commentIDs = append(commentIDs, ids...)
	}

	commentIDs = dedupeCommentIDs(commentIDs)
	if len(commentIDs) == 0 {
		return &CleanupResult{}, nil
	}

	replyCountAdjustments, err := buildReplyCountAdjustments(tx, commentIDs)
	if err != nil {
		return nil, err
	}

	result := &CleanupResult{}
	err = tx.Transaction(func(trx *gorm.DB) error {
		deletedLikeCount, deleteLikeErr := deleteCommentLikes(trx, commentIDs)
		if deleteLikeErr != nil {
			return deleteLikeErr
		}
		result.DeletedLikeCount = deletedLikeCount

		deleteCommentsResult := trx.Where("id IN ?", commentIDs).Delete(&ds.PluginComment{})
		if deleteCommentsResult.Error != nil {
			return deleteCommentsResult.Error
		}
		result.DeletedCommentCount = deleteCommentsResult.RowsAffected

		for rootID, delta := range replyCountAdjustments {
			if delta <= 0 {
				continue
			}
			if updateErr := trx.Model(&ds.PluginComment{}).
				Where("id = ?", rootID).
				UpdateColumn(
					"reply_cnt",
					gorm.Expr("GREATEST(reply_cnt - ?, 0)", delta),
				).Error; updateErr != nil {
				return updateErr
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LikeComment 连续点赞一次，底层复用 act_like 的三次上限能力。
func LikeComment(req LikeRequest, user *security.JwtUser) (*LikeResult, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}
	if user == nil || user.Id == 0 {
		return nil, errors.New("请先登录")
	}
	id, err := parseUintID(req.Id)
	if err != nil {
		return nil, err
	}
	var comment ds.PluginComment
	if err := tx.First(&comment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("评论不存在")
		}
		return nil, err
	}
	bizType := uint8(enum.PluginComment)
	if err := (&active.Like{Id: id, BizType: bizType}).Add(user); err != nil {
		return nil, err
	}
	total, liked, err := getLikeState(tx, id, user.Id)
	if err != nil {
		return nil, err
	}
	return &LikeResult{Id: strconv.FormatUint(id, 10), LikeCnt: total, LikedByMe: liked}, nil
}

func cleanupAllInstanceComments(tx *gorm.DB, instanceID string) (*CleanupResult, error) {
	var commentIDs []uint64
	err := tx.Model(&ds.PluginComment{}).
		Where("instance_id = ?", instanceID).
		Pluck("id", &commentIDs).Error
	if err != nil {
		return nil, err
	}
	commentIDs = dedupeCommentIDs(commentIDs)
	if len(commentIDs) == 0 {
		return &CleanupResult{}, nil
	}

	result := &CleanupResult{}
	err = tx.Transaction(func(trx *gorm.DB) error {
		deletedLikeCount, deleteLikeErr := deleteCommentLikes(trx, commentIDs)
		if deleteLikeErr != nil {
			return deleteLikeErr
		}
		result.DeletedLikeCount = deletedLikeCount

		deleteCommentsResult := trx.Where("id IN ?", commentIDs).Delete(&ds.PluginComment{})
		if deleteCommentsResult.Error != nil {
			return deleteCommentsResult.Error
		}
		result.DeletedCommentCount = deleteCommentsResult.RowsAffected
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func buildCommentViews(tx *gorm.DB, comments []ds.PluginComment, user *security.JwtUser) ([]CommentView, error) {
	if len(comments) == 0 {
		return []CommentView{}, nil
	}
	userMap := loadUsers(tx, comments)
	likeMap, likedMap, err := loadLikeMaps(tx, comments, user)
	if err != nil {
		return nil, err
	}
	views := make([]CommentView, 0, len(comments))
	for _, comment := range comments {
		anchor := CommentAnchor{}
		if comment.AnchorJson != "" {
			_ = json.Unmarshal([]byte(comment.AnchorJson), &anchor)
		}
		view := CommentView{
			Id:              strconv.FormatUint(comment.Id, 10),
			Anchor:          anchor,
			Level:           comment.Level,
			Author:          authorView(comment.CreatedBy, userMap[comment.CreatedBy]),
			Content:         comment.Content,
			LikeCnt:         likeMap[comment.Id],
			ReplyCnt:        comment.ReplyCnt,
			LikedByMe:       likedMap[comment.Id],
			ReplyToNickname: comment.ReplyToNickname,
			CreatedOn:       comment.CreatedAt.Format(time.RFC3339Nano),
		}
		if !comment.UpdatedAt.IsZero() {
			view.UpdatedOn = comment.UpdatedAt.Format(time.RFC3339Nano)
		}
		if comment.ParentId != nil {
			view.ParentId = strconv.FormatUint(*comment.ParentId, 10)
		}
		if comment.RootId != nil {
			view.RootId = strconv.FormatUint(*comment.RootId, 10)
		}
		views = append(views, view)
	}
	return views, nil
}

func loadUsers(tx *gorm.DB, comments []ds.PluginComment) map[uint64]sys.User {
	ids := make([]uint64, 0, len(comments))
	seen := map[uint64]bool{}
	for _, comment := range comments {
		if comment.CreatedBy == 0 || seen[comment.CreatedBy] {
			continue
		}
		seen[comment.CreatedBy] = true
		ids = append(ids, comment.CreatedBy)
	}
	users := map[uint64]sys.User{}
	if len(ids) == 0 {
		return users
	}
	var list []sys.User
	if err := tx.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return users
	}
	for _, user := range list {
		users[user.Id] = user
	}
	return users
}

func loadLikeMaps(tx *gorm.DB, comments []ds.PluginComment, user *security.JwtUser) (map[uint64]int64, map[uint64]bool, error) {
	bizType := uint8(enum.PluginComment)
	likeMap := map[uint64]int64{}
	likedMap := map[uint64]bool{}
	items := make([]active.Like, 0, len(comments))
	ids := make([]uint64, 0, len(comments))
	for _, comment := range comments {
		items = append(items, active.Like{Id: comment.Id, BizType: bizType})
		ids = append(ids, comment.Id)
	}
	counts, err := active.GetBatchCounts(items)
	if err != nil {
		return nil, nil, err
	}
	for _, count := range counts {
		likeMap[count.Id] = count.Total
	}
	if user == nil || user.Id == 0 || len(ids) == 0 {
		return likeMap, likedMap, nil
	}
	var likes []active.Like
	if err := tx.Where("user_id = ? AND biz_type = ? AND data_id IN ?", user.Id, bizType, ids).Find(&likes).Error; err != nil {
		return nil, nil, err
	}
	for _, like := range likes {
		likedMap[like.Id] = like.Num > 0
	}
	return likeMap, likedMap, nil
}

func getLikeState(tx *gorm.DB, id uint64, userId uint64) (int64, bool, error) {
	counts, err := active.GetBatchCounts([]active.Like{{Id: id, BizType: uint8(enum.PluginComment)}})
	if err != nil {
		return 0, false, err
	}
	total := int64(0)
	if len(counts) > 0 {
		total = counts[0].Total
	}
	var like active.Like
	err = tx.Where("data_id = ? AND user_id = ? AND biz_type = ?", id, userId, uint8(enum.PluginComment)).First(&like).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, err
	}
	return total, like.Num > 0, nil
}

func authorView(userId uint64, user sys.User) AuthorView {
	nickname := strings.TrimSpace(user.Nickname)
	if nickname == "" {
		nickname = fmt.Sprintf("用户%d", userId)
	}
	return AuthorView{
		Id:         strconv.FormatUint(userId, 10),
		Nickname:   nickname,
		AvatarText: firstRune(nickname),
		City:       strings.TrimSpace(user.City),
	}
}

func marshalAnchor(anchor CommentAnchor) (string, error) {
	if len(anchor) == 0 {
		return "", errors.New("评论锚点不能为空")
	}
	bytes, err := json.Marshal(anchor)
	if err != nil {
		return "", errors.New("评论锚点格式错误")
	}
	return string(bytes), nil
}

func anchorString(anchor CommentAnchor, key string) string {
	if anchor == nil {
		return ""
	}
	if value, ok := anchor[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func anchorIndex(anchor CommentAnchor) (int, bool) {
	if anchor == nil {
		return 0, false
	}
	value, ok := anchor["index"]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case int:
		return typed, true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		return parsed, err == nil
	default:
		return 0, false
	}
}

func anchorItemID(anchor CommentAnchor) string {
	return anchorString(anchor, "itemId")
}

func listCommentIDsByItemID(tx *gorm.DB, instanceID string, itemID string) ([]uint64, error) {
	var ids []uint64
	err := tx.Model(&ds.PluginComment{}).
		Where("instance_id = ? AND anchor_item_id = ?", instanceID, itemID).
		Pluck("id", &ids).Error
	return ids, err
}

func listCommentIDsByIndex(tx *gorm.DB, instanceID string, index int) ([]uint64, error) {
	var ids []uint64
	err := tx.Model(&ds.PluginComment{}).
		Where("instance_id = ? AND anchor_index = ?", instanceID, index).
		Pluck("id", &ids).Error
	return ids, err
}

func listCommentSubtreeIDsByCommentID(
	tx *gorm.DB,
	instanceID string,
	commentID string,
	userID uint64,
) ([]uint64, error) {
	id, err := parseUintID(commentID)
	if err != nil {
		return nil, err
	}

	var comments []ds.PluginComment
	if err := tx.Where("instance_id = ?", instanceID).Find(&comments).Error; err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return []uint64{}, nil
	}

	commentByID := make(map[uint64]ds.PluginComment, len(comments))
	childrenByParent := make(map[uint64][]uint64)
	for _, comment := range comments {
		commentByID[comment.Id] = comment
		if comment.ParentId != nil {
			childrenByParent[*comment.ParentId] = append(childrenByParent[*comment.ParentId], comment.Id)
		}
	}

	target, ok := commentByID[id]
	if !ok {
		return nil, errors.New("评论不存在")
	}
	if target.CreatedBy != userID {
		return nil, errors.New("只能删除自己的评论")
	}

	ids := []uint64{id}
	queue := []uint64{id}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		children := childrenByParent[current]
		if len(children) == 0 {
			continue
		}
		ids = append(ids, children...)
		queue = append(queue, children...)
	}
	return dedupeCommentIDs(ids), nil
}

// 回复数量挂在根评论上，删除回复树时需要同步扣减。
func buildReplyCountAdjustments(tx *gorm.DB, commentIDs []uint64) (map[uint64]int, error) {
	if len(commentIDs) == 0 {
		return map[uint64]int{}, nil
	}

	var comments []ds.PluginComment
	if err := tx.Where("id IN ?", commentIDs).Find(&comments).Error; err != nil {
		return nil, err
	}

	deletedSet := make(map[uint64]struct{}, len(commentIDs))
	for _, id := range commentIDs {
		deletedSet[id] = struct{}{}
	}

	adjustments := make(map[uint64]int)
	for _, comment := range comments {
		if comment.RootId == nil {
			continue
		}
		if _, deletingRoot := deletedSet[*comment.RootId]; deletingRoot {
			continue
		}
		adjustments[*comment.RootId]++
	}
	return adjustments, nil
}

func dedupeCommentIDs(ids []uint64) []uint64 {
	if len(ids) <= 1 {
		return ids
	}
	seen := make(map[uint64]struct{}, len(ids))
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func deleteCommentLikes(tx *gorm.DB, commentIDs []uint64) (int64, error) {
	if len(commentIDs) == 0 {
		return 0, nil
	}
	deleteResult := tx.
		Where("biz_type = ? AND data_id IN ?", uint8(enum.PluginComment), commentIDs).
		Delete(&active.Like{})
	if deleteResult.Error != nil {
		return 0, deleteResult.Error
	}
	return deleteResult.RowsAffected, nil
}

func normalizePage(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func parseUintID(raw string) (uint64, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("id无效")
	}
	return id, nil
}

func firstRune(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return ""
	}
	return string(runes[0])
}
