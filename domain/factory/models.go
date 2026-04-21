package factory

// 迁移表列表。
func Tables() []interface{} {
	return []interface{}{
		&Release{},
		&ReleasePriceHistory{},
		&ReleaseStatusHistory{},
		&MintRecord{},
		&UserOwnership{},
		&UpgradeRecord{},
	}
}
