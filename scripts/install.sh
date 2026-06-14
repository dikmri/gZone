#!/bin/sh
# gZone インストールスクリプト (macOS / Linux)
# 使い方: curl -fsSL https://raw.githubusercontent.com/dikmri/gZone/main/scripts/install.sh | sh

set -e

REPO="dikmri/gZone"
INSTALL_DIR="/usr/local/bin"

# OS / アーキテクチャ検出
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  GOOS="linux" ;;
  Darwin) GOOS="darwin" ;;
  *)
    echo "エラー: 未対応のOS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)  GOARCH="amd64" ;;
  aarch64|arm64) GOARCH="arm64" ;;
  *)
    echo "エラー: 未対応のアーキテクチャ: $ARCH"
    exit 1
    ;;
esac

# 最新タグを取得
echo "最新バージョンを確認中..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "エラー: バージョン情報を取得できませんでした"
  exit 1
fi

echo "バージョン: $LATEST"

ASSET="gZone_${GOOS}_${GOARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"

# macOS の場合は .app.tar.gz を優先（初回インストール時は .app バンドルを使用）
if [ "$GOOS" = "darwin" ]; then
  echo "ダウンロード中: ${ASSET}.app.tar.gz ..."
  TMP_DIR=$(mktemp -d)
  curl -fsSL "${URL}.app.tar.gz" -o "${TMP_DIR}/gZone.app.tar.gz"
  tar xzf "${TMP_DIR}/gZone.app.tar.gz" -C "${TMP_DIR}"

  APP_DEST="/Applications/gZone.app"
  if [ -d "$APP_DEST" ]; then
    echo "既存の gZone.app を削除中..."
    rm -rf "$APP_DEST"
  fi
  cp -r "${TMP_DIR}/gZone.app" /Applications/
  xattr -d com.apple.quarantine /Applications/gZone.app 2>/dev/null || true
  rm -rf "$TMP_DIR"
  echo "インストール完了: /Applications/gZone.app"
  echo "Launchpad または Finder から起動できます。"
else
  # Linux: バイナリを /usr/local/bin に配置
  echo "ダウンロード中: ${ASSET} ..."
  TMP=$(mktemp)
  curl -fsSL "$URL" -o "$TMP"
  chmod +x "$TMP"

  if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "${INSTALL_DIR}/gZone"
  else
    echo "管理者権限が必要です (sudo)"
    sudo mv "$TMP" "${INSTALL_DIR}/gZone"
  fi

  echo "インストール完了: ${INSTALL_DIR}/gZone"
  echo "ターミナルで 'gZone' を実行して起動できます。"
fi
