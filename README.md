# gZone

ローカルの動画ファイルをサムネイルグリッドで閲覧できるデスクトップアプリです。  
Go + Wails v2 で作られており、Windows / macOS / Linux に対応しています。

## 特徴

- フォルダを選択するだけで動画を自動スキャン
- グリッド表示 + サイズスライダーで一覧を自由に調整
- スクロールに連動した遅延ロード・自動解放（大量ファイルでも軽快）
- ホバーで動画を自動再生（音声なし）
- 起動時に自動アップデートチェック

## インストール

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/dikmri/gZone/main/scripts/install.ps1 | iex
```

`%LOCALAPPDATA%\gZone\gZone.exe` にインストールされ、スタートメニューにショートカットが作成されます。

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/dikmri/gZone/main/scripts/install.sh | sh
```

- **macOS**: `/Applications/gZone.app` にインストールされます
- **Linux**: `/usr/local/bin/gZone` にインストールされます

### 手動ダウンロード

[Releases](https://github.com/dikmri/gZone/releases) ページから最新バージョンをダウンロードしてください。

| プラットフォーム | ファイル |
|---|---|
| Windows (x64) | `gZone_windows_amd64.exe` |
| macOS (Apple Silicon) | `gZone_darwin_arm64.app.tar.gz` |
| macOS (Intel) | `gZone_darwin_amd64.app.tar.gz` |
| Linux (x64) | `gZone_linux_amd64` |

## 使い方

1. アプリを起動する
2. 「フォルダを選択」ボタンで動画フォルダを開く
3. スライダーでサムネイルサイズを調整する
4. 動画にホバーすると自動再生

## 動作環境

- **Windows**: Windows 10 / 11 (WebView2 が必要、通常は自動インストール済み)
- **macOS**: macOS 12 Monterey 以降
- **Linux**: GTK3 + WebKit2GTK が必要 (`libgtk-3` / `libwebkit2gtk-4.0`)

## 自動アップデート

起動時に GitHub Releases を確認し、新バージョンがある場合は自動でダウンロード・適用・再起動します。

## ビルド方法

```bash
# 依存パッケージのインストール
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# 開発モード（ホットリロード）
wails dev

# リリースビルド
wails build -ldflags "-X main.Version=v0.1.0"
```

## ライセンス

MIT
