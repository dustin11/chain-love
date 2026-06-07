package ds

import "senspace/domain"

// PluginComment 插件通用评论。
type PluginComment struct {
	Id              uint64  `json:"id,string" gorm:"primaryKey;autoIncrement;comment:评论ID"`
	InstanceId      string  `json:"instanceId" gorm:"type:varchar(128);not null;index:idx_plugin_comment_instance_created,priority:1;index:idx_plugin_comment_instance_index,priority:1;comment:插件实例ID"`
	AnchorIndex     *int    `json:"anchorIndex,omitempty" gorm:"index:idx_plugin_comment_instance_index,priority:2;comment:锚点索引"`
	AnchorJson      string  `json:"anchorJson" gorm:"type:json;comment:完整锚点JSON"`
	ParentId        *uint64 `json:"parentId,string,omitempty" gorm:"index:idx_plugin_comment_parent;comment:父评论ID"`
	RootId          *uint64 `json:"rootId,string,omitempty" gorm:"index:idx_plugin_comment_root;comment:根评论ID"`
	Level           int     `json:"level" gorm:"not null;default:0;comment:评论层级"`
	ReplyToNickname string  `json:"replyToNickname,omitempty" gorm:"type:varchar(128);comment:被回复昵称"`
	Content         string  `json:"content" gorm:"type:text;not null;comment:评论内容"`
	ReplyCnt        int     `json:"replyCnt" gorm:"not null;default:0;comment:回复数"`
	domain.CreatInfo
	domain.UpdateInfo
}

// TableName 表名。
func (PluginComment) TableName() string {
	return "ds_plugin_comment"
}
