package active

import (
	"chain-love/domain"
	"chain-love/pkg/app/security"
	"chain-love/pkg/e"
	"errors"

	"gorm.io/gorm"
)

type Like struct {
	Id      uint64 `json:"id" gorm:"column:data_id;primaryKey;autoIncrement:false;comment:数据ID"`
	UserId  uint64 `json:"userId" gorm:"primaryKey;autoIncrement:false;comment:用户ID"`
	BizType uint8  `json:"bizType" gorm:"primaryKey;autoIncrement:false;comment:业务类型"`
	Num     uint8  `json:"num" gorm:"comment:数量"`
	domain.CreatInfo
}

func (Like) TableName() string {
	return "act_like"
}

type LikeCountVO struct {
	Id    uint64 `json:"id"`
	Total int64  `json:"total"`
}

// Add 点赞：存在则累加（上限3），不存在则创建
func (m *Like) Add(user *security.JwtUser) error {
	m.UserId = user.Id
	m.CreatedBy = user.Id

	var old Like
	err := domain.Db.Where("data_id = ? AND user_id = ? AND biz_type = ?", m.Id, m.UserId, m.BizType).First(&old).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.Num = 1
			return domain.Db.Create(m).Error
		}
		return err
	}

	e.PanicIfBizErr(old.Num >= 3, "", 1001)

	return domain.Db.Model(&old).Update("num", old.Num+1).Error
}

// Delete 删除点赞记录
func (m *Like) Delete(user *security.JwtUser) error {
	return domain.Db.Where("data_id = ? AND user_id = ? AND biz_type = ?", m.Id, user.Id, m.BizType).Delete(&Like{}).Error
}

// GetBatchCounts 批量统计点赞数
func GetBatchCounts(items []Like) ([]LikeCountVO, error) {
	if len(items) == 0 {
		return nil, nil
	}
	// 构建 OR 查询条件： (data_id = ? AND biz_type = ?) OR ...
	db := domain.Db.Table("act_like").Select("data_id as id, sum(num) as total")

	var logicStr string
	var args []interface{}
	for i, item := range items {
		if i > 0 {
			logicStr += " OR "
		}
		logicStr += "(data_id = ? AND biz_type = ?)"
		args = append(args, item.Id, item.BizType)
	}

	var results []LikeCountVO
	err := db.Where(logicStr, args...).Group("data_id, biz_type").Scan(&results).Error
	return results, err
}
