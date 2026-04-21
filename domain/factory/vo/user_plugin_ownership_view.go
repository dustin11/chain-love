package vo

import "senspace/domain/factory"

// 用户插件资产视图。
type UserPluginOwnershipView struct {
	Id                       string                        `json:"id"`
	UserId                   string                        `json:"userId"`
	PluginId                 string                        `json:"pluginId"`
	MintedReleaseId          string                        `json:"mintedReleaseId"`
	EffectiveReleaseId       string                        `json:"effectiveReleaseId"`
	CreatedAt                string                        `json:"createdAt"`
	UpdatedAt                string                        `json:"updatedAt"`
	UpgradedAt               string                        `json:"upgradedAt,omitempty"`
	MintedVersion            string                        `json:"mintedVersion,omitempty"`
	EffectiveVersion         string                        `json:"effectiveVersion,omitempty"`
	LatestAvailableReleaseId string                        `json:"latestAvailableReleaseId,omitempty"`
	LatestAvailableVersion   string                        `json:"latestAvailableVersion,omitempty"`
	UpgradeState             factory.OwnershipUpgradeState `json:"upgradeState"`
	UpgradePrice             string                        `json:"upgradePrice,omitempty"`
}
