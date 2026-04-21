# 数字工厂 dev 环境测试。
# 测试代码会按服务启动路径建表，导入 asset/database 下的 SQL 测试数据，再执行测试。
# 测试完成后保留表和数据，便于复测与联调。

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$oldSpiderEnv = $env:SPIDER_ENV
$oldGoCache = $env:GOCACHE

try {
    Set-Location $repoRoot
    $env:SPIDER_ENV = "dev"
    $env:GOCACHE = Join-Path $repoRoot ".gocache"
    go test ./test -run 'TestFactory' -count=1
}
finally {
    Set-Location $repoRoot
    if ($null -eq $oldSpiderEnv -or $oldSpiderEnv -eq "") {
        Remove-Item Env:SPIDER_ENV -ErrorAction SilentlyContinue
    }
    else {
        $env:SPIDER_ENV = $oldSpiderEnv
    }

    if ($null -eq $oldGoCache -or $oldGoCache -eq "") {
        Remove-Item Env:GOCACHE -ErrorAction SilentlyContinue
    }
    else {
        $env:GOCACHE = $oldGoCache
    }
}
