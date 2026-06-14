# gZone インストールスクリプト (Windows)
# 使い方: irm https://raw.githubusercontent.com/dikmri/gZone/main/scripts/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo    = "dikmri/gZone"
$AppName = "gZone"
$InstallDir = Join-Path $env:LOCALAPPDATA "gZone"

Write-Host "最新バージョンを確認中..."

$ApiUrl  = "https://api.github.com/repos/$Repo/releases/latest"
$Release = Invoke-RestMethod -Uri $ApiUrl -UseBasicParsing
$Version = $Release.tag_name

if (-not $Version) {
    Write-Error "バージョン情報を取得できませんでした"
    exit 1
}

Write-Host "バージョン: $Version"

$Asset   = "gZone_windows_amd64.exe"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Asset"

# インストール先ディレクトリを作成
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force $InstallDir | Out-Null
}

$Dest = Join-Path $InstallDir "$AppName.exe"

Write-Host "ダウンロード中: $Asset ..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $Dest -UseBasicParsing

Write-Host "インストール完了: $Dest"

# PATH に追加（ユーザーレベル）
$CurrentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    [System.Environment]::SetEnvironmentVariable(
        "PATH",
        "$CurrentPath;$InstallDir",
        "User"
    )
    Write-Host "PATH に追加しました: $InstallDir"
    Write-Host "新しいターミナルを開いてから 'gZone' コマンドが使えます。"
} else {
    Write-Host "PATH はすでに設定されています。"
}

# スタートメニューにショートカットを作成
$StartMenu = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs"
$Shortcut  = Join-Path $StartMenu "gZone.lnk"
$WshShell  = New-Object -ComObject WScript.Shell
$Link      = $WshShell.CreateShortcut($Shortcut)
$Link.TargetPath       = $Dest
$Link.WorkingDirectory = $InstallDir
$Link.Description      = "gZone - 動画ギャラリー"
$Link.Save()

Write-Host "スタートメニューにショートカットを作成しました。"
Write-Host ""
Write-Host "gZone を起動するには:"
Write-Host "  - スタートメニューから 'gZone' を検索して起動"
Write-Host "  - またはコマンド 'gZone' を実行"
