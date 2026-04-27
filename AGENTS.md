# Copilot 使用指南

本仓库当前以后端 Go 代码为主，生成或修改代码时，必须优先遵守本文件中的 Go 代码规范与编码安全规范。

---

# 1. 任务执行规则

这些规则控制 Copilot 如何执行任务。

- 始终确保代码可编译、可运行。
- 不允许引入 `go build` / `go test` 层面的错误。
- 保持现有项目分层与目录结构不被随意破坏。
- 除非明确要求，否则仅修改完成任务所必需的最少代码。
- 不要擅自修改文件编码、换行风格或已有 API 语义。
- 仔细执行每项任务，不遗漏文档中已经明确的需求点。

编码规则：

- 所有源码文件必须保持 UTF-8 编码。
- 注释必须支持中文字符，且不能出现乱码。
- 提交前应保证新增和修改的 Go 文件可被 `gofmt` 正常格式化。

错误处理规则：

- 不允许忽略错误。
- 错误必须按仓库现有风格明确处理，不允许 silent failure。
- 若需要向前端返回提示信息，必须使用统一错误封装，不要自行拼装响应结构。

---

# 2. Go 通用编码规则

语言规则：

- 优先使用 **Go** 实现后端逻辑。
- 结构体、方法、函数、请求体、响应体命名要清晰、稳定、可读。
- 能复用现有类型时，不重复创建语义重叠的新类型。

风格规则：

- 使用描述性的变量名，避免无意义缩写。
- 函数职责保持单一，避免一个函数同时处理参数解析、业务编排、持久化和响应输出。
- 避免出现大段重复代码，优先抽取共用函数。
- 减少深层嵌套，优先使用提前返回让主流程更清晰。
- 先遵循仓库既有风格，再考虑个人偏好。

接口分层规则：

- `api` 层负责参数接收、调用 service、统一响应。
- `service` 层负责业务编排、校验、跨实体逻辑。
- `domain` 层负责模型定义、数据库操作和领域对象行为。
- 不要把大量业务逻辑直接堆在 `api` 层。

注释规则：
- 公共接口、导出类型、枚举、复杂业务逻辑必须有中文注释, 注释风格简洁重点。
- 注释只描述重点，避免在注释开头重复类型名、方法名、字段名或把标识符翻译一遍。
- 注释风格优先参考 [`auth_api.go`](/f:/project/senspace/senspace/api/api_v1/auth_api/auth_api.go)。
- Swagger 注释、普通中文注释、关键逻辑前的短注释可以混合使用，但必须保持简洁、直接、可读。
- 除非注释本身错误或已过期，否则不要删除已有有效注释。
- Go 中不要写 JSDoc 风格要求，改为使用 Go 注释风格和 Swagger 注释风格。

注释示例：

```go
// @Summary 获取登录 Nonce
// @Tags 认证
// @Param addr path string true "用户addr"
// @Success 200 {object} e.Error
// @Router /api/v1/auth/nonce/{addr} [get]
func GetNonce(c *gin.Context) {
	addr := c.Param("addr")
	// 参数校验改为统一抛出参数错误
	e.PanicIfParameterError(addr == "", "地址不能为空")
	e.PanicIfParameterError(!strings.HasPrefix(addr, "0x") || len(addr) != 42, "无效的钱包地址")

	nonce := auth_service.GenerateNonce(strings.ToLower(addr))
	app.Response(c, e.SuccessData(map[string]string{"nonce": nonce}))
}
```

结构体与请求体示例：

```go
// 登录请求体。
type LoginReq struct {
	Message   string `json:"message" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}
```

---

# 3. API 与错误处理规范

本项目 API 风格应优先参考：

- [`auth_api.go`](/f:/project/senspace/senspace/api/api_v1/auth_api/auth_api.go)
- [`plugin_api.go`](/f:/project/senspace/senspace/api/api_v1/dev_api/plugin_api.go)

## 1. API 层基本规范

- `gin.Context` 或 `contextx.AppContext` 负责接收请求参数。
- 参数绑定后立即校验，错误要立刻按统一方式抛出。
- API 层不直接返回裸 JSON，不自行构造 `{code,msg,data}`。
- 成功响应统一使用：
  - `app.Response(c, e.Success)`
  - `app.Response(c, e.SuccessData(data))`

示例：

```go
func AddFolder(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
		Path     string `json:"path"`
	}
	err := c.ShouldBindJSON(&req)
	e.PanicParameterError(err)

	err = dev_service.AddFolder(req.PluginId, req.Path)
	e.PanicServerErr(err)

	app.Response(c, e.Success)
}
```

## 2. 参数错误处理规范

参数错误必须使用 `pkg/e` 中已有方法，禁止手写杂散错误风格。

优先使用的错误类型：

- `e.PanicIfParameterError(condition, message)`
  - 用于主动条件判断失败。
- `e.PanicParameterError(err)`
  - 用于 `ShouldBindJSON`、`ShouldBind` 等绑定失败。
- `e.PanicParameterErrorTipMsg(err, message)`
  - 用于需要给出明确参数提示时，例如文件缺失、表单非法。

示例：

```go
file, err := c.FormFile("file")
e.PanicParameterErrorTipMsg(err, "file missing")
```

```go
if err := c.ShouldBindJSON(&req); err != nil {
	e.PanicIfParameterError(err != nil, "参数错误")
}
```

## 3. 服务端错误处理规范

内部错误、IO 错误、数据库错误、服务调用错误，统一使用已有封装：

- `e.PanicServerErr(err)`
- `e.PanicServerErrTipMsg(err, message)`
- `e.PanicIfErr(err)`
  - 仅在符合现有项目语义时使用，避免滥用。

使用建议：

- 默认优先 `e.PanicServerErr(err)`。
- 需要给前端稳定业务提示时，使用 `e.PanicServerErrTipMsg(err, "自定义提示")`。
- 不要同时记录日志又重复包装多层同义错误，保持错误链路简单。

示例：

```go
tree, err := dev_service.GetPluginTree(pluginId)
e.PanicServerErr(err)
app.Response(c, e.SuccessData(tree))
```

```go
err = dev_service.Rename(req.PluginId, req.OldPath, req.NewName)
e.PanicServerErrTipMsg(err, "重命名失败")
app.Response(c, e.Success)
```

## 4. 允许的错误处理风格

- 允许 service 内部继续返回 `error`，由 API 层统一转成 `e.*`。
- 允许 service 内部在项目已有模式下直接 `panic`，但新代码优先保持边界清晰。
- 不允许：
  - 忽略 `err`
  - `fmt.Println(err)` 后继续执行
  - 自行 `c.JSON(...)` 输出非统一格式错误
  - 同一层混用多套风格

## 5. 响应风格规范

- 查询成功并返回数据：`app.Response(c, e.SuccessData(data))`
- 仅表示成功：`app.Response(c, e.Success)`
- 使用 `contextx.AppContext` 时，响应对象使用 `c.Gin`

示例：

```go
func SavePlugin(c *contextx.AppContext) {
	pluginId := c.Gin.Query("pluginId")
	form, err := c.Gin.MultipartForm()
	e.PanicParameterErrorTipMsg(err, "invalid multipart form")

	res, err := dev_service.SavePlugin(c, pluginId, form)
	e.PanicServerErr(err)
	app.Response(c.Gin, e.SuccessData(res))
}
```

---

# 4. 代码设计原则

在编写、修改或重构 Go 代码时，必须遵循以下设计原则。

## 1. 保持模块单一职责

每个模块、类型或函数应只负责一个明确功能。

错误示例：

- 一个 API 方法同时处理参数绑定、数据库读写、文件系统操作和复杂业务决策。

正确示例：

- API 负责接收请求
- Service 负责业务编排
- Domain 负责模型与持久化

---

## 2. 保持代码结构清晰

避免复杂嵌套和难以理解的逻辑。

规则：

- 函数应尽量短小，主流程一眼能读懂。
- 嵌套层级尽量不超过 3 层。
- 多分支逻辑优先拆成小函数。
- 优先使用早返回处理异常路径。

---

## 3. 优先复用现有能力

实现功能前，优先检查仓库中是否已有可复用的方法、领域对象或错误封装。

规则：

- 先扩展现有模块，再新增模块。
- 先复用已有错误处理函数，再新增错误类型。
- 先复用已有响应格式，再新增输出结构。

---

## 4. 明确数据流

数据流应清晰可追踪，避免隐式状态。

规则：

- 避免隐藏副作用。
- 参数、返回值和状态变更要明确。
- 不要让调用方猜测函数是否会写库、改状态或触发响应。

---

## 5. 可测试性

代码应易于测试。

规则：

- 避免把核心逻辑写死在 HTTP Handler 中。
- 尽量让 service 能被直接测试。
- 避免无边界地依赖全局状态。

---

## 6. 错误处理必须明确

所有可能失败的操作都必须处理错误。

规则：

- 数据库操作必须检查错误。
- 文件操作必须检查错误。
- 参数绑定必须检查错误。
- 不允许 silent failure。

---

## 7. 不允许重复代码（DRY）

重复逻辑必须提取为函数或模块。

规则：

- 如果出现两次以上相似参数解析、错误处理或响应逻辑，应优先考虑抽取。

---

## 8. 新类型添加原则

如果需要添加新结构体、请求体、响应体或领域模型，先分析现有类型是否可扩展。

规则：

- 仅在已有类型不能满足需求时新增类型。
- 新类型必须语义清晰，命名准确。
- 新增类型时，要说明其职责边界。

---

# 5. 测试流程

这些规则用于约束自动化测试，避免污染现有开发数据和历史功能。

- 自动化测试默认复用 `asset/conf/dev.yml` 作为配置来源。
- 默认直接连接 dev 库执行测试，但测试范围必须严格限制在当前功能保留的测试数据内，禁止影响其它功能数据、表结构和历史记录。
- 若测试依赖建表，必须走与 Go 服务重启一致的 GORM 初始化路径，例如 `domain.Setup` + `d_util.InitTable`，禁止手工改动正式表结构。
- 测试数据必须以 `.sql` 文件形式放在 `asset/database` 目录。完整流程为：读取 dev 配置、执行建表初始化、导入测试数据、执行测试方法、校验结果。
- 测试数据必须使用“保留测试数据范围”设计，例如固定测试 ID、测试前缀、专用插件 ID、专用业务标记等，确保操作范围可控、可复跑、可识别。
- 如需在重复测试前重置数据，必须通过 `asset/database` 下的 `.sql` 文件仅重置当前功能的保留测试数据，禁止删表、清空其它业务表或改动其它功能数据。
- 测试脚本统一放在 `asset/database` 目录，脚本中必须体现“dev 配置来源”“建表流程”“SQL 导入”“测试执行”“保留测试数据范围”。
- 测试使用到的插件目录、缓存目录必须使用临时路径，不得写入现有业务目录。
- 除非用户明确要求清理，测试完成后默认保留测试表和测试数据，作为后续复测、排查和联调用例的依据。
- 测试结束后，必须恢复测试中临时改写的全局配置或环境变量。
- 如果无法确认数据库操作只作用于当前功能的保留测试数据，必须先停止执行，再调整测试方案。

---

# 6. 编码安全硬门槛

这是最高优先级规则之一，优先级高于一般开发效率。

## 1. 修改前必做

- 任何读取、修改、创建源码文件前，必须先按 UTF-8 环境执行。
- 在 Windows PowerShell 5.1 中，禁止直接使用未显式声明编码的 `Get-Content` / `gc` / `cat` 读取源码文件。
- 读取 UTF-8 无 BOM 源码时，必须显式使用 `Get-Content -Encoding utf8`，或使用 `[System.IO.File]::ReadAllText/ReadAllLines(path, [System.Text.UTF8Encoding]::new($false, $true))`。
- 仅设置控制台 `InputEncoding` / `OutputEncoding` 不足以保证源码读取正确，这只能修复终端收发编码，不能修复 `Get-Content` 的默认解码。
- 必须先确认目标文件是否存在乱码、BOM、异常注释编码问题。
- 如果终端看到中文乱码，必须先用“显式 UTF-8 重读 + 字节检查”复核，确认是真文件乱码还是 PowerShell 默认解码导致的假乱码；未复核前，不得把文件判定为已损坏。
- 如果目标文件编码状态不明确，不允许直接开始功能修改。

## 2. 发现乱码时的强制流程

- 一旦发现目标文件或相邻代码存在乱码，立即停止功能改动。
- 先恢复该文件的编码安全，再继续功能修改。
- 禁止在乱码文本附近直接做局部 patch。
- 优先使用“从稳定版本重建局部代码块”或“整段注释重写”的方式处理乱码。

## 3. 修改中的限制

- 不允许把功能修改和乱码修复混在同一个高风险 patch 里。
- 如果文件编码状态不确定，不允许继续扩大修改范围。
- 注释、字符串、文档必须保持 UTF-8 正常中文，不允许出现乱码占位内容。

## 4. 修改后必做

- 对本次 touched files 逐一检查：
  - UTF-8
  - 无 BOM
  - 无 mojibake pattern
- 如果任一文件未通过，必须先修复，不能结束任务。
- 编码检查未通过时，禁止宣称任务完成。

## 5. 输出要求

- 每次任务结束时，必须明确汇报：
  - 哪些文件做了编码检查
  - 是否 UTF-8 无 BOM
  - 是否发现乱码
- 如果为了保证编码安全而暂停了功能修改，也必须明确说明。
