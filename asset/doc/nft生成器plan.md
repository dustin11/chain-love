# 通用 NFT 生成器执行计划

本文档规划后端 Go 版通用 NFT 生成器、发布后 metadata 生成、预铸造登记表和库存队列表。目标是让 FishTank 鱼 NFT 从“前端生成 JSON + mint 时临时分配”的 MVP，升级为“发布时后端冻结集合 + 导入通用库存池 + mint 时事务发放”的通用市场能力。

## 核心结论

- 正式发行的 NFT 预生成放在后端更合适，前端生成器保留为调参、预览和抽检工具。
- `asset.meta.json` 统一使用 `metadataRef` 描述 NFT 元数据来源，不再兼容旧字段 `ref`。
- `metadataRef` 可以按 collection `key` 和 `tierConfig` 推导库存来源，不额外增加 `inventory.source`。
- 所有预生成固定集合型 NFT 统一导入通用库存池，不做 FishTank 专用库存表。
- mint 时只从库存表领取 `available` item，并在同一事务内写入资产、铸造记录和库存状态，避免并发重复发放。

## asset.meta.json 约定

### 新字段

`collections[]` 使用以下字段：

```json
{
  "label": "鱼",
  "key": "fish",
  "assetKind": "fish",
  "metadataRef": "generated/fish/{tier}.json",
  "traitHashField": "traitHash",
  "tierConfig": {
    "common": { "price": "5", "supply": 51000, "mintLimit": 10 },
    "rare": { "price": "50", "supply": 8400, "mintLimit": 5 },
    "epic": { "price": "500", "supply": 588, "mintLimit": 1 },
    "legendary": { "price": "-", "supply": 12, "mintLimit": 0 }
  }
}
```

普通模板型 NFT：

```json
{
  "label": "鱼缸",
  "key": "tank",
  "assetKind": "tank",
  "metadataRef": "defaultWaterMeta.json#tanks",
  "unitPrice": "10"
}
```

### 固定字段规则

- `metadataRef` 必填，表示 NFT item 的元数据来源。
- `key` 必填，作为 collection 业务键，例如 `fish`、`tank`、`avatar`、`weapon`。
- item 的 ID 字段固定为 `id`，不提供 `idField` 配置。
- item 的等级字段固定为 `tier`，不提供 `tierField` 配置。
- item 的序号从 1 开始，`indexBase` 固定为 1。
- `traitHashField` 可选；如果配置，则 item 必须包含该字段。
- 如果 collection 配置了 `tierConfig`，则每个 item 必须包含 `tier`，且值必须命中 `tierConfig`。
- 如果 `metadataRef` 包含 `{tier}`，后端按 `tierConfig` 的 key 展开读取。
- 如果 `metadataRef` 不包含 `{tier}`，后端按普通 JSON ref 读取。
- 不兼容旧字段 `ref`。发布校验遇到缺少 `metadataRef` 时直接报错。

## Go 版通用 NFT 生成器

Go 版生成器不是单个 FishTank 函数，而是一组通用发布后任务：

```text
release snapshot ready
  ↓
read asset.meta.json
  ↓
resolve metadataRef for each collection
  ↓
generate or load item metadata
  ↓
validate schema / tier / id / traitHash
  ↓
write frozen metadata JSON
  ↓
create inventory pool and inventory items
  ↓
build collection hash / Merkle root
  ↓
mark release inventory ready
```

### 生成器类型

第一阶段支持两类：

- `template`：从 `metadataRef` 直接读取模板数组，例如 `defaultWaterMeta.json#tanks`。
- `pregenerated`：从已经生成好的固定集合读取，例如 `generated/fish/{tier}.json`。

后续可以扩展：

- `backend-generator`：后端根据配置和 seed 直接生成 item，FishTank Go 版生成器属于这个类型。
- `external-import`：从外部 JSON、CSV 或后台上传文件导入。
- `on-demand`：不提前写满库存，mint 时动态生成或绑定。

### FishTank Go 版生成器改进点

FishTank 正式发行建议迁到后端生成：

- Go 读取生成配置、色板、纹理 archetype、性格规则和 legendary 固定配置。
- 使用固定 seed PRNG，不依赖前端运行时。
- 生成后立即运行 schema、颜色、复杂度、traitHash 和重复校验。
- 生成结果写入 release 静态快照，而不是先写前端工作区。
- 生成摘要、校验报告、collection hash 和 config hash 一并写入快照。
- 生成完成后直接导入库存池，形成可审计发行状态。

前端生成器继续作为“实验室”：

- 调参。
- 批量预览。
- 抽检。
- 生成候选配置。
- 展示概率分布。

## 通用库存池数据模型

### InventoryPool

建议新增表：`fact_nft_inventory_pool`。

```go
type NFTInventoryPool struct {
    Id             int64
    PluginId       string
    ReleaseId      int64
    CollectionKey  string
    AssetKind      factory.AssetKind
    MetadataRef    string
    Strategy       string
    TotalSupply    int64
    MintedCount    int64
    Status         string
    CollectionHash string
    MerkleRoot     string
    GeneratedAt    *time.Time
    FrozenAt       *time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

推荐状态：

- `draft`：已创建，未冻结。
- `frozen`：item 已导入，集合不可变。
- `active`：可 mint。
- `exhausted`：库存已发完。
- `closed`：关闭。

推荐发放策略：

- `shuffled`：预生成集合按私有打乱顺序发放，适合 FishTank 鱼。
- `sequential`：按序发放。
- `auction`：拍卖或人工分配，适合 legendary。
- `allowDuplicate`：模板允许重复 mint。
- `onDemand`：mint 时动态生成。

### InventoryItem

建议新增表：`fact_nft_inventory_item`。

```go
type NFTInventoryItem struct {
    Id            int64
    PoolId        int64
    PluginId      string
    ReleaseId     int64
    CollectionKey string
    AssetKind     factory.AssetKind

    ItemId        string
    ItemIndex     int64
    Tier          string
    TraitHash     string
    ShuffleIndex  int64
    MetadataHash  string
    LeafHash      string
    ProofJson     string

    Status        string
    ReservedBy    string
    ReservedAt    *time.Time
    ReservedUntil *time.Time

    AssetId       *int64
    TokenId       string
    MintRecordId  *int64
    OwnerKey      string
    MintedAt      *time.Time

    MetadataUri   string
    ProofUri      string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

推荐 item 状态：

- `available`：可发放。
- `reserved`：临时锁定，等待支付或链上确认。
- `minted`：已发放。
- `burned`：已销毁。
- `voided`：作废，不再发放。

### 索引和约束

建议约束：

- `(release_id, collection_key)` 唯一对应一个 pool，或按 tier 拆 pool 时 `(release_id, collection_key, tier)` 唯一。
- `(pool_id, item_id)` 唯一。
- `(release_id, collection_key, item_id)` 唯一。
- `(release_id, asset_id)` 可选唯一，确保 item 和资产绑定唯一。

建议索引：

- `pool_id + status + tier + shuffle_index`：mint 时取库存。
- `release_id + collection_key + tier + status`：查剩余库存。
- `owner_key + status`：用户库存关系查询。
- `trait_hash`：审计和去重。

## 发布后 metadata 生成

### 通用 metadata builder

发布后读取每个 item，生成市场标准 metadata：

```json
{
  "name": "FishTank Fish fish-common-000001",
  "description": "FishTank composable fish NFT.",
  "animation_url": "/static/factory/releases/FishTank/1.0.0-910000000000001/release.json",
  "attributes": [
    { "trait_type": "Asset Kind", "value": "fish" },
    { "trait_type": "Collection", "value": "fish" },
    { "trait_type": "Tier", "value": "common" },
    { "trait_type": "Trait Hash", "value": "d2828e991c9529e7" }
  ],
  "properties": {
    "pluginId": "FishTank",
    "releaseId": "910000000000001",
    "collectionKey": "fish",
    "itemId": "fish-common-000001",
    "itemIndex": 1,
    "metadataRef": "generated/fish/common.json",
    "traitHash": "d2828e991c9529e7"
  }
}
```

FishTank 可以额外映射：

- `paletteId` -> `Palette`
- `archetypeId` 或 `themePreset` -> `Pattern` / `Theme`
- `finArchetypeId` -> `Fin Pattern`
- `finStyle` -> `Fin Style`
- `dorsalFinShape` -> `Dorsal Fin`
- `personalityId` -> `Personality`
- `color` -> `Body Color`
- `tailColor` -> `Tail Color`
- `eyeColor` -> `Eye Color`

### 静态文件输出

发布后预生成：

```text
metadata/by-item/{collectionKey}/{itemId}.json
proofs/by-item/{collectionKey}/{itemId}.json
metadata/collection/{collectionKey}.json
```

mint 后绑定 token：

```text
metadata/{tokenId}.json
proofs/{tokenId}.json
assets/{assetId}.json
```

如果 metadata 需要包含 tokenId，则 mint 时从 by-item metadata 派生 token metadata。

## Merkle Root 与可验证性

发布后为每个 pool 构建 Merkle tree。

推荐 leaf 内容：

```text
releaseId | collectionKey | itemIndex | itemId | tier | traitHash | metadataHash
```

流程：

1. 生成 item metadata。
2. 计算 `metadataHash`。
3. 计算 leaf hash。
4. 构建 Merkle root。
5. `merkleRoot` 写入 inventory pool。
6. 每个 item 保存 proof。
7. mint 后 token proof 从 item proof 派生。

第一阶段不上链时，`merkleRoot` 存 DB 和 release 静态 manifest。后续上链时再写入合约或发布事件。

## Mint 发放流程

通用 mint 流程：

```text
1. 前端提交 collectionKey + tier + count
2. 后端读取 asset.meta.json 和 inventory pool
3. 校验 tierConfig、价格、mintLimit、pool 状态
4. 事务内锁定 release 和 pool
5. 按 strategy 取 available item
6. 标记 item 为 reserved 或 minted
7. 创建 fact_mint_record
8. 创建 fact_asset
9. 回写 item.assetId / tokenId / mintRecordId / ownerKey / mintedAt
10. 写 metadata/{tokenId}.json 和 proofs/{tokenId}.json
11. 写 owner index 和 composition snapshot
```

推荐 SQL 形态：

```sql
SELECT *
FROM fact_nft_inventory_item
WHERE pool_id = ?
  AND tier = ?
  AND status = 'available'
ORDER BY shuffle_index
LIMIT ?
FOR UPDATE;
```

同一事务内更新：

```sql
UPDATE fact_nft_inventory_item
SET status = 'minted',
    asset_id = ?,
    token_id = ?,
    mint_record_id = ?,
    owner_key = ?,
    minted_at = ?
WHERE id = ?;
```

## 发布校验

发布或构建 ready 前必须校验：

- `asset.meta.json` schema 合法。
- 每个 collection 有 `key`、`assetKind`、`metadataRef`。
- `metadataRef` 文件存在，且路径不能越界。
- item 必须有 `id`。
- item `id` 在同 collection 内唯一。
- 如果有 `tierConfig`，item 必须有 `tier`，且 `tier` 必须命中配置。
- 如果 `metadataRef` 包含 `{tier}`，必须配置 `tierConfig`。
- 如果配置 `traitHashField`，item 必须包含该字段。
- `tierConfig.supply` 必须等于对应 tier item 数量。
- 禁用价格 `"-"` 的 tier 可以导入库存，但默认不可 mint。
- `mintLimit` 为 0 的 tier 不可选。
- collection hash 和 Merkle root 可重算一致。

## FishTank 迁移计划

### 阶段 A：asset.meta.json 改造

把 FishTank 改成统一 `metadataRef`：

```json
{
  "label": "鱼缸",
  "key": "tank",
  "assetKind": "tank",
  "metadataRef": "defaultWaterMeta.json#tanks",
  "unitPrice": "10"
}
```

```json
{
  "label": "鱼",
  "key": "fish",
  "assetKind": "fish",
  "metadataRef": "generated/fish/{tier}.json",
  "traitHashField": "traitHash",
  "tierConfig": {
    "common": { "price": "5", "supply": 51000, "mintLimit": 10 },
    "rare": { "price": "50", "supply": 8400, "mintLimit": 5 },
    "epic": { "price": "500", "supply": 588, "mintLimit": 1 },
    "legendary": { "price": "-", "supply": 12, "mintLimit": 0 }
  }
}
```

### 阶段 B：通用 metadataRef 读取

- 后端 `assetValueCollection` 去掉 `Ref`，新增 `MetadataRef`。
- `resolveMintSelection` 按 `metadataRef` 找 collection。
- 前端 mint 弹窗使用 `metadataRef` 作为输入 key，或者使用 `key` 作为输入 key。推荐最终使用 `key`，减少路径暴露。
- 所有旧 `ref` 逻辑删除，不保留兼容。

### 阶段 C：库存池建模

- 新增 `NFTInventoryPool` 和 `NFTInventoryItem` domain model。
- 加入 `factory.Tables()` 自动迁移。
- 发布后任务根据 `asset.meta.json` 创建 pool 和 items。
- FishTank 鱼导入时按 tier 读取 `generated/fish/{tier}.json`。
- 为每个 item 生成 `shuffle_index`。

### 阶段 D：mint 改为库存表发放

- `MintReleaseAsset` 不再扫描 JSON 和 `fact_asset` 去重。
- mint 时锁定 pool 和 items。
- 发放成功后回写 item 和 asset。
- 库存不足直接报错。
- owner snapshot 和 composition snapshot 保持现有逻辑。

### 阶段 E：Go 版 FishTank 生成器

- 将 TypeScript 配置 JSON 迁移到后端可读取目录。
- Go 实现 seed PRNG、权重抽样、颜色规则、traitHash、校验规则。
- 发布构建时生成 FishTank 鱼集合。
- 生成结果写入 release snapshot。
- 前端生成器只作为预览工具，不再作为正式发行源。

## 与当前 MVP 的差异

当前 MVP：

```text
前端 generated/fish/*.json
  ↓
release snapshot 复制 JSON
  ↓
mint 时扫描 JSON + 查 fact_asset 去重
  ↓
写资产 metadata/proof
```

目标形态：

```text
后端生成或导入 item
  ↓
发布后冻结 metadataRef item
  ↓
导入 fact_nft_inventory_pool / fact_nft_inventory_item
  ↓
生成 metadata、proof、Merkle root
  ↓
mint 时从库存表事务发放
  ↓
写 fact_asset、token metadata、owner snapshot
```

目标形态的收益：

- 发行集合可审计。
- 库存状态显式。
- 并发发放更安全。
- 支持 reserved、超时释放和链上确认。
- 支持拍卖、白名单、盲盒 reveal。
- 适用于 FishTank 之外的头像、装备、票券、徽章等 NFT。

## 已决策实现取向

- 前端 mint 请求使用 `collection.key` 作为输入 key，不暴露 `metadataRef` 路径。
- pool 第一阶段按 collection 建一条，item 上保留 `tier`。
- metadata 第一阶段先在导入库存时计算 by-item hash，mint 时派生 by-token metadata/proof；后续补 by-item 静态文件批量输出。
- FishTank legendary 进入同一 `fish` collection；因为 `price: "-"` 且 `mintLimit: 0`，默认不可公开 mint，后续拍卖流程再单独接入。
