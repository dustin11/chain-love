package plugin_comment_service

// CommentAnchor 插件评论锚点，服务端原样保存插件传入的结构。
type CommentAnchor map[string]any

// ListRequest 评论分页筛选参数。
type ListRequest struct {
	InstanceId string        `json:"instanceId"`
	Mode       string        `json:"mode"`
	Anchor     CommentAnchor `json:"anchor"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
}

// CreateRequest 新建评论参数。
type CreateRequest struct {
	Anchor          CommentAnchor `json:"anchor"`
	Content         string        `json:"content"`
	ParentId        string        `json:"parentId"`
	ReplyToNickname string        `json:"replyToNickname"`
}

// LikeRequest 评论点赞参数。
type LikeRequest struct {
	Id string `json:"id"`
}

// AuthorView 评论作者展示数据。
type AuthorView struct {
	Id         string `json:"id"`
	Nickname   string `json:"nickname"`
	AvatarText string `json:"avatarText,omitempty"`
	City       string `json:"city,omitempty"`
}

// CommentView 前端评论展示数据。
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

// ListResult 评论分页结果。
type ListResult struct {
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Items    []CommentView `json:"items"`
}

// LikeResult 点赞后的评论点赞状态。
type LikeResult struct {
	Id        string `json:"id"`
	LikeCnt   int64  `json:"likeCnt"`
	LikedByMe bool   `json:"likedByMe"`
}
