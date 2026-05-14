package task_service

import (
	"errors"
	"os"
	"path/filepath"

	"senspace/domain/factory"
	"senspace/domain/task"
	"senspace/pkg/setting"

	"gorm.io/gorm"
)

// PurgeAllMintData 清空全部铸造资产、任务记录与工厂静态资源。
func PurgeAllMintData() error {
	if !setting.IsDevLikeEnv() {
		return errors.New("仅开发环境支持清空全部任务")
	}
	tx, err := db()
	if err != nil {
		return err
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM fact_asset_relation").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM fact_asset").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM fact_mint_record").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM fact_user_ownership").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM fact_nft_inventory_item").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM fact_nft_inventory_pool").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE fact_release SET minted_count = 0").Error; err != nil {
			return err
		}
		if err := tx.Where("task_type IN ?", []task.Type{task.TypeMint, task.TypePublish}).
			Delete(&task.AsyncTask{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return purgeFactoryStaticFiles()
}

// purgeFactoryStaticFiles 清空工厂静态目录下的铸造产物，不影响发布快照。
func purgeFactoryStaticFiles() error {
	root := factory.FactoryStaticRoot()
	for _, name := range []string{
		"assets",
		"metadata",
		"proofs",
		"owners",
	} {
		if err := os.RemoveAll(filepath.Join(root, name)); err != nil {
			return err
		}
	}
	return nil
}
