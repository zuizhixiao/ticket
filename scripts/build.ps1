# ============ 电影纪念票根 · 一键构建 ============
# 用法: .\scripts\build.ps1          # 构建 UI 并产出二进制 ticket.exe
#       $env:SKIP_UI='1' 跳过 UI 构建

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if ($env:SKIP_UI -ne '1') {
    Write-Host '[1/2] 构建前端(web → internal/assets/dist)...'
    Push-Location web
    try {
        if (-not (Test-Path node_modules)) { npm install --ignore-scripts }
        npm run build
    } finally {
        Pop-Location
    }
} else {
    Write-Host '[1/2] 跳过前端构建(SKIP_UI=1)'
}

Write-Host '[2/2] 编译后端二进制 ticket.exe...'
$env:CGO_ENABLED = '0'
go build -o ticket.exe ./cmd/server
Write-Host "完成: $root\ticket.exe"
Write-Host '本地运行: RUN_MODE=dev 时读取 config.yaml;生产请配置环境变量后执行 ./ticket.exe'
