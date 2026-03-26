# CODEX 执行规则

## 编码规则

1. 任何包含中文的源码、文档、配置文件，统一使用 UTF-8 无 BOM 编码保存。
2. 使用 `shell_command` 写文件前，终端必须先执行以下设置：

```powershell
chcp 65001 > $null
[Console]::InputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
```

3. 使用脚本写文件时，必须显式指定 UTF-8 无 BOM，例如：

```powershell
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($path, $content, $utf8NoBom)
```

4. 不允许依赖 PowerShell 默认编码，也不允许直接把含中文内容通过未声明编码的重定向写入文件。

## 修改规则

1. 优先使用 `apply_patch` 修改文件。
2. 如果 `apply_patch` 不可用，改用 `shell_command` 时必须遵守本文件中的编码规则。
3. 修改历史文件时，先确认原文件编码正常，再追加或替换内容。
4. 对大文件做脚本替换时，优先按结构定位，不要仅依赖中文注释文本匹配。

## 校验规则

1. 修改含中文文件后，至少做一次 UTF-8 内容校验。
2. 校验时不要只看终端显示，要按字节读取文件并确认内容字符串正常。
3. 如果发现典型乱码（如“鎻掍欢”“闂”“鍦”这类由 UTF-8 被误按 GBK 解读后的文本），应使用可逆修复方式处理。

## 乱码修复规则

1. 若文件出现“UTF-8 被误按 GBK 解读”型乱码，优先按行执行可逆修复：
   - 先 `encode('gbk')`
   - 再 `decode('utf-8')`
2. 修复时必须只替换可成功恢复且明显更合理的行，避免伤及原本正常的中文。
3. 修复完成后，重新以 UTF-8 无 BOM 写回文件。

## 验证规则

1. 代码修改后优先执行 `npm run ts`。
2. 如需进一步确认，再执行 `npm run build`。
3. 如果构建因沙箱权限失败，应说明是环境权限问题，不应误判为源码错误。