@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo ========================================
echo  网络日志格式化系统 - 构建脚本
echo ========================================
echo.

:: 检查 Go 环境
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go，请先安装 Go: https://go.dev/dl/
    exit /b 1
)

:: 检查 Wails CLI
where wails >nul 2>&1
if %errorlevel% neq 0 (
    echo [提示] 未找到 Wails CLI，正在安装...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if %errorlevel% neq 0 (
        echo [错误] Wails CLI 安装失败
        exit /b 1
    )
    echo [完成] Wails CLI 已安装
)

:: 整理依赖
echo [1/2] 整理依赖...
go mod tidy
if %errorlevel% neq 0 (
    echo [错误] go mod tidy 失败
    exit /b 1
)
echo [完成] 依赖已整理
echo.

:: 构建
echo [2/2] 构建应用...
wails build -clean -ldflags "-H windowsgui"
if %errorlevel% neq 0 (
    echo [错误] 构建失败
    exit /b 1
)

echo.
echo ========================================
echo  构建成功！
echo  输出: build\bin\
echo ========================================
