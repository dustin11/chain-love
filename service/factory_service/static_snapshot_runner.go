package factory_service

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"senspace/domain/factory"
)

const timeLayoutSecond = "2006-01-02T15:04:05"

// 立即重建发布静态快照。
func rebuildReleaseStaticSnapshotNow(release factory.Release) error {
	stageDir, err := factory.StageReleaseStaticSnapshot(release)
	if err != nil {
		return err
	}

	backupDir, err := factory.ActivateReleaseStaticSnapshot(release, stageDir)
	if err != nil {
		_ = os.RemoveAll(stageDir)
		return err
	}

	if err := factory.CommitActivatedReleaseStaticSnapshot(backupDir); err != nil {
		rollbackErr := factory.RollbackActivatedReleaseStaticSnapshot(release, backupDir)
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	return nil
}

// 立即重建持有人的资产索引和组合快照。
func rebuildOwnerFactorySnapshotsNow(ownerKey string) error {
	tx, err := db()
	if err != nil {
		return err
	}

	var assets []factory.Asset
	if err := tx.
		Where("owner_key = ? AND status = ?", ownerKey, factory.AssetStatusActive).
		Order("created_at DESC").
		Find(&assets).Error; err != nil {
		return err
	}

	var relations []factory.AssetRelation
	if err := tx.
		Where("owner_key = ? AND status = ?", ownerKey, factory.AssetRelationStatusActive).
		Order("created_at ASC").
		Find(&relations).Error; err != nil {
		return err
	}

	stageDir, err := stageOwnerFactorySnapshots(ownerKey, assets, relations)
	if err != nil {
		return err
	}
	return activateOwnerFactorySnapshots(ownerKey, stageDir)
}

func stageOwnerFactorySnapshots(
	ownerKey string,
	assets []factory.Asset,
	relations []factory.AssetRelation,
) (string, error) {
	stageDir := filepath.Dir(factory.OwnerIndexStaticPath(ownerKey)) + ".staging"
	if err := os.RemoveAll(stageDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(stageDir, 0755); err != nil {
		return "", err
	}

	now := time.Now().Format(time.RFC3339Nano)
	index := ownerFactoryAssetIndex{
		Schema:    "senspace.factory.owner-assets.v2",
		OwnerKey:  ownerKey,
		UpdatedAt: now,
		Assets:    make([]ownerFactoryAssetEntry, 0, len(assets)),
	}
	for _, asset := range assets {
		if err := writeFactoryAssetSnapshot(asset); err != nil {
			_ = os.RemoveAll(stageDir)
			return "", err
		}
		index.Assets = append(index.Assets, ownerFactoryAssetEntry{
			AssetId:       strconv.FormatInt(asset.Id, 10),
			PluginId:      asset.PluginId,
			ReleaseId:     strconv.FormatInt(asset.ReleaseId, 10),
			Version:       asset.Version,
			RuntimeKind:   defaultRuntimeKind(asset.RuntimeKind),
			AssetKind:     asset.AssetKind,
			CollectionKey: asset.CollectionKey,
			ComponentRole: asset.ComponentRole,
			ParentKey:     asset.ParentKey,
			TemplateRef:   asset.TemplateRef,
			ItemId:        asset.ItemId,
			ItemIndex:     asset.ItemIndex,
			PluginOptions: decodePluginOptions(asset.PluginOptions),
			Tier:          asset.Tier,
			TraitHash:     asset.TraitHash,
			AssetUrl:      factory.AssetStaticURL(asset.PluginId, asset.Id),
			ReleaseUrl:    releaseStaticURLFromAsset(asset),
			MintedAt:      asset.CreatedAt.Format(time.RFC3339Nano),
		})
	}
	if err := factory.WriteJSONAtomic(filepath.Join(stageDir, "assets.json"), index); err != nil {
		_ = os.RemoveAll(stageDir)
		return "", err
	}

	composition := ownerFactoryAssetComposition{
		Schema:    "senspace.factory.owner-composition.v1",
		OwnerKey:  ownerKey,
		UpdatedAt: now,
		Relations: make([]ownerFactoryAssetCompositionEdge, 0, len(relations)),
	}
	for _, relation := range relations {
		composition.Relations = append(composition.Relations, ownerFactoryAssetCompositionEdge{
			Id:            strconv.FormatInt(relation.Id, 10),
			RelationType:  relation.RelationType,
			SourceAssetId: strconv.FormatInt(relation.SourceAssetId, 10),
			TargetAssetId: strconv.FormatInt(relation.TargetAssetId, 10),
			Metadata:      decodeRelationMetadata(relation.MetadataJson),
		})
	}
	if err := factory.WriteJSONAtomic(filepath.Join(stageDir, "composition.json"), composition); err != nil {
		_ = os.RemoveAll(stageDir)
		return "", err
	}
	return stageDir, nil
}

func activateOwnerFactorySnapshots(ownerKey string, stageDir string) error {
	finalDir := filepath.Dir(factory.OwnerIndexStaticPath(ownerKey))
	backupDir := finalDir + ".previous"
	if err := os.MkdirAll(filepath.Dir(finalDir), 0755); err != nil {
		return err
	}
	if err := os.RemoveAll(backupDir); err != nil {
		return err
	}
	if info, err := os.Stat(finalDir); err == nil && info.IsDir() {
		if err := os.Rename(finalDir, backupDir); err != nil {
			return err
		}
	}
	if err := os.Rename(stageDir, finalDir); err != nil {
		if _, statErr := os.Stat(backupDir); statErr == nil {
			_ = os.Rename(backupDir, finalDir)
		}
		return err
	}
	return os.RemoveAll(backupDir)
}
