#!/bin/bash
#
# publish.sh — 一键构建并打包 beryllium-be-there 为可发布目录
#
# 用法:
#   ./publish.sh              # 构建所有组件并打包到 publish/ 目录
#   ./publish.sh clean        # 清理构建产物
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

PUBLISH_DIR="$SCRIPT_DIR/publish"
SERVER_DIR="$SCRIPT_DIR/beryllium-server"
MANAGE_DIR="$SCRIPT_DIR/beryllium-manage"
SIGNIN_DIR="$SCRIPT_DIR/beryllium-signin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ---------------------------------------------------------------------------
# clean — 清理所有构建产物
# ---------------------------------------------------------------------------
clean() {
    log_info "清理构建产物..."

    rm -rf "$PUBLISH_DIR"
    rm -rf "$MANAGE_DIR/out"
    rm -rf "$MANAGE_DIR/.next"
    rm -rf "$SIGNIN_DIR/out"
    rm -rf "$SIGNIN_DIR/.next"
    rm -f  "$SERVER_DIR/beryllium-be-there"

    log_info "清理完成"
}

if [ "${1:-}" = "clean" ]; then
    clean
    exit 0
fi

# ---------------------------------------------------------------------------
# 1. 构建管理端前端
# ---------------------------------------------------------------------------
log_info "构建 beryllium-manage..."
cd "$MANAGE_DIR"

if [ ! -d "node_modules" ]; then
    log_info "安装依赖..."
    npm install
fi

npm run build
log_info "beryllium-manage 构建完成 → $MANAGE_DIR/out"

# ---------------------------------------------------------------------------
# 2. 构建签到端前端
# ---------------------------------------------------------------------------
log_info "构建 beryllium-signin..."
cd "$SIGNIN_DIR"

if [ ! -d "node_modules" ]; then
    log_info "安装依赖..."
    npm install
fi

npm run build
log_info "beryllium-signin 构建完成 → $SIGNIN_DIR/out"

# ---------------------------------------------------------------------------
# 3. 构建 Go 后端
# ---------------------------------------------------------------------------
log_info "构建 beryllium-server..."
cd "$SERVER_DIR"

# 确定目标平台
GOOS="${GOOS:-}"
GOARCH="${GOARCH:-}"
if [ -z "$GOOS" ]; then
    GOOS=$(go env GOOS)
fi
if [ -z "$GOARCH" ]; then
    GOARCH=$(go env GOARCH)
fi

BINARY_NAME="beryllium-be-there"
if [ "$GOOS" = "windows" ]; then
    BINARY_NAME="beryllium-be-there.exe"
fi

# 编译
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -ldflags="-s -w" \
    -o "$BINARY_NAME" \
    .

log_info "beryllium-server 构建完成 → $SERVER_DIR/$BINARY_NAME ($GOOS/$GOARCH)"

# ---------------------------------------------------------------------------
# 4. 打包到 publish 目录
# ---------------------------------------------------------------------------
log_info "打包到 $PUBLISH_DIR..."

rm -rf "$PUBLISH_DIR"
mkdir -p "$PUBLISH_DIR"

# 复制可执行文件
cp "$SERVER_DIR/$BINARY_NAME" "$PUBLISH_DIR/"

# 复制前端构建产物
mkdir -p "$PUBLISH_DIR/manage"
cp -r "$MANAGE_DIR/out/"* "$PUBLISH_DIR/manage/"

mkdir -p "$PUBLISH_DIR/signin"
cp -r "$SIGNIN_DIR/out/"* "$PUBLISH_DIR/signin/"

# 默认以当前目录下的 manage/ 和 signin/ 为静态文件目录
# 与生产部署时的目录结构一致

log_info "打包完成!"

# ---------------------------------------------------------------------------
# 5. 输出发布目录结构
# ---------------------------------------------------------------------------
echo ""
echo "============================================"
echo "  发布目录: $PUBLISH_DIR"
echo "  目标平台: $GOOS/$GOARCH"
echo "============================================"
echo ""
echo "目录结构:"
find "$PUBLISH_DIR" -type f | sed "s|$PUBLISH_DIR|publish|" | sort
echo ""
echo "运行方法:"
echo "  cd publish"
echo "  ./$BINARY_NAME"
echo ""
echo "访问地址:"
echo "  管理端: http://localhost:8080/"
echo "  签到端: http://localhost:8080/signin/{classId}"
echo ""
echo "默认管理员: admin / admin"
