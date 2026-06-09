package plugin_comment_service

// 评论锚点，服务端原样保存插件传入的结构。
type CommentAnchor map[string]any

// 评论分页筛选参数。
type ListRequest struct {
	InstanceId string        `json:"instanceId"`
	Mode       string        `json:"mode"`
	Anchor     CommentAnchor `json:"anchor"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
}

// 新建评论参数。
type CreateRequest struct {
	Anchor          CommentAnchor `json:"anchor"`
	Content         string        `json:"content"`
	ParentId        string        `json:"parentId"`
	ReplyToNickname string        `json:"replyToNickname"`
}

// 评论级联清理参数。
type CleanupRequest struct {
	InstanceId string        `json:"instanceId"`
	DeleteAll  bool          `json:"deleteAll,omitempty"`
	Items      []CleanupItem `json:"items"`
}

// 评论级联清理目标。
type CleanupItem struct {
	CommentId string `json:"commentId,omitempty"`
	ItemId    string `json:"itemId,omitempty"`
	Index     *int   `json:"index,omitempty"`
}

// 评论点赞参数。
type LikeRequest struct {
	Id string `json:"id"`
}

// 评论作者展示数据。
type AuthorView struct {
	Id         string `json:"id"`
	Nickname   string `json:"nickname"`
	AvatarText string `json:"avatarText,omitempty"`
	City       string `json:"city,omitempty"`
}

// 前端评论展示数据。
type CommentView struct {
	Id              string        `json:"id"`
	Anchor          CommentAnchor `json:"anchor"`
	ParentId        string        `json:"parentId,omitempty"`
	RootId          string        `json:"rootId,omitempty"`
	Level           int           `json:"level"`
	Author          AuthorView    `json:"author"`
	Content         string        `json:"content"`
	LikeCnt         int64         `json:"likeCnt"`
	ReplyCnt        int           `json:"replyCnt"`
	LikedByMe       bool          `json:"likedByMe"`
	ReplyToNickname string        `json:"replyToNickname,omitempty"`
	CreatedOn       string        `json:"createdOn"`
	UpdatedOn       string        `json:"updatedOn,omitempty"`
}

// 评论分页结果。
type ListResult struct {
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Items    []CommentView `json:"items"`
}

// 点赞后的评论状态。
type LikeResult struct {
	Id        string `json:"id"`
	LikeCnt   int64  `json:"likeCnt"`
	LikedByMe bool   `json:"likedByMe"`
}

// 评论级联清理结果。
type CleanupResult struct {
	DeletedCommentCount int64 `json:"deletedCommentCount"`
	DeletedLikeCount    int64 `json:"deletedLikeCount"`
}
