#!/bin/bash
#
# publish.sh — 一键构建并打包 beryllium-be-there 为多平台可发布目录
#
# 用法:
#   ./publish.sh                  # 构建所有平台
#   ./publish.sh darwin/amd64     # 仅构建指定平台
#   ./publish.sh clean            # 清理构建产物
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

PUBLISH_DIR="$SCRIPT_DIR/publish"
SERVER_DIR="$SCRIPT_DIR/beryllium-server"
MANAGE_DIR="$SCRIPT_DIR/beryllium-manage"
SIGNIN_DIR="$SCRIPT_DIR/beryllium-signin"

# 默认构建目标平台
DEFAULT_PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step()  { echo -e "${CYAN}[STEP]${NC} $1"; }

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
    rm -f  "$SERVER_DIR/beryllium-be-there" "$SERVER_DIR/beryllium-be-there.exe"
    log_info "清理完成"
}

# ---------------------------------------------------------------------------
# build_frontend — 构建前端，平台无关，只需构建一次
# ---------------------------------------------------------------------------
build_frontend() {
    local name="$1"
    local dir="$2"

    log_info "构建 $name..."
    cd "$dir"

    if [ ! -d "node_modules" ]; then
        log_info "  → 安装依赖..."
        npm install --silent
    fi

    npm run build --silent
    log_info "  → 构建完成: $dir/out"
}

# ---------------------------------------------------------------------------
# build_go — 交叉编译 Go 后端
# ---------------------------------------------------------------------------
build_go() {
    local goos="$1"
    local goarch="$2"
    local output="$3"

    cd "$SERVER_DIR"

    log_info "  → 编译 $goos/$goarch ..."
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
        -ldflags="-s -w" \
        -o "$output" \
        .
    log_info "  → 完成: $output"
}

# ---------------------------------------------------------------------------
# package — 将构建产物打包到一个平台目录
# ---------------------------------------------------------------------------
package() {
    local goos="$1"
    local goarch="$2"
    local binary_name="$3"

    local platform_dir="$PUBLISH_DIR/$goos-$goarch"
    mkdir -p "$platform_dir"

    # 复制可执行文件
    cp "$SERVER_DIR/$binary_name" "$platform_dir/"

    # 复制前端构建产物
    cp -r "$MANAGE_DIR/out" "$platform_dir/manage"
    cp -r "$SIGNIN_DIR/out" "$platform_dir/signin"

    # 复制配置文件
    cp "$SERVER_DIR/config.toml" "$platform_dir/"

    log_info "  → 打包完成: $platform_dir"
}

# ---------------------------------------------------------------------------
# print_summary — 输出发布包摘要
# ---------------------------------------------------------------------------
print_summary() {
    echo ""
    echo "============================================"
    echo "  发布目录: $PUBLISH_DIR"
    echo "============================================"
    echo ""
    echo "构建目标:"
    for dir in "$PUBLISH_DIR"/*/; do
        local name
        name=$(basename "$dir")
        local size
        size=$(du -sh "$dir" 2>/dev/null | cut -f1)
        printf "  %-25s %s\n" "$name" "$size"
    done
    echo ""
    echo "每个发布包内包含:"
    echo "  ├── beryllium-be-there[.exe]   # 可执行文件"
    echo "  ├── config.toml                # 配置文件"
    echo "  ├── manage/                    # 管理端前端"
    echo "  └── signin/                    # 签到端前端"
    echo ""
    echo "运行方法:"
    echo "  cd publish/<platform>"
    echo "  ./beryllium-be-there           # Linux / macOS"
    echo "  beryllium-be-there.exe         # Windows"
    echo ""
    echo "访问地址:"
    echo "  管理端: http://{host}:8080/"
    echo "  签到端: http://{host}:8080/signin/{classId}"
    echo ""
    echo "默认管理员: admin / admin"
}

# ============================================================================
# main
# ============================================================================

if [ "${1:-}" = "clean" ]; then
    clean
    exit 0
fi

# 确定要构建的平台列表
if [ $# -gt 0 ]; then
    PLATFORMS=("$@")
else
    PLATFORMS=("${DEFAULT_PLATFORMS[@]}")
fi

# 验证平台格式
for p in "${PLATFORMS[@]}"; do
    if [[ ! "$p" =~ ^[a-z]+/[a-z0-9_]+$ ]]; then
        log_error "无效的平台格式: $p (应为 os/arch，如 darwin/amd64)"
        exit 1
    fi
done

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║     beryllium-be-there 构建发布脚本      ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "目标平台: ${#PLATFORMS[@]} 个"
for p in "${PLATFORMS[@]}"; do
    echo "  - $p"
done
echo ""

# ---------------------------------------------------------------------------
# Step 1: 构建前端（平台无关，构建一次）
# ---------------------------------------------------------------------------
log_step "[1/3] 构建前端"
build_frontend "beryllium-manage" "$MANAGE_DIR"
build_frontend "beryllium-signin" "$SIGNIN_DIR"
echo ""

# ---------------------------------------------------------------------------
# Step 2: 清理旧发布目录
# ---------------------------------------------------------------------------
log_step "[2/3] 准备发布目录"
rm -rf "$PUBLISH_DIR"
mkdir -p "$PUBLISH_DIR"
echo ""

# ---------------------------------------------------------------------------
# Step 3: 交叉编译并打包
# ---------------------------------------------------------------------------
log_step "[3/4] 交叉编译 Go 后端并打包"

COUNT=0
TOTAL=${#PLATFORMS[@]}

for platform in "${PLATFORMS[@]}"; do
    COUNT=$((COUNT + 1))
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"

    binary="beryllium-be-there"
    if [ "$GOOS" = "windows" ]; then
        binary="beryllium-be-there.exe"
    fi

    log_info "[$COUNT/$TOTAL] 构建 $GOOS/$GOARCH"

    build_go "$GOOS" "$GOARCH" "$binary"
    package "$GOOS" "$GOARCH" "$binary"
    echo ""
done

# 清理编译中间产物
rm -f "$SERVER_DIR/$binary" "$SERVER_DIR/beryllium-be-there.exe" 2>/dev/null || true

# ---------------------------------------------------------------------------
# Step 4: 压缩为 zip 文件
# ---------------------------------------------------------------------------
log_step "[4/4] 压缩发布包"
ZIP_FILE="$SCRIPT_DIR/beryllium-be-there.zip"
rm -f "$ZIP_FILE"
cd "$PUBLISH_DIR" && zip -r "$ZIP_FILE" ./* && cd "$SCRIPT_DIR"
log_info "压缩完成: $ZIP_FILE"
echo ""

# ---------------------------------------------------------------------------
# 输出摘要
# ---------------------------------------------------------------------------
print_summary
