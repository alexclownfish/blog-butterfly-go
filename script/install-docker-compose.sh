#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" 2>/dev/null && pwd || pwd)"
ROOT_DIR_DEFAULT="$(cd "${SCRIPT_DIR}/.." 2>/dev/null && pwd || pwd)"
ROOT_DIR="${ROOT_DIR:-$ROOT_DIR_DEFAULT}"
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
PUBLIC_HOST="${PUBLIC_HOST:-127.0.0.1}"
BUILD_FLAG="--build"
ACTION="up"
INSTALL_DOCKER_IF_MISSING=1
REPO_URL="${REPO_URL:-https://github.com/alexclownfish/blog-butterfly-go.git}"
REPO_REF="${REPO_REF:-main}"
INSTALL_DIR="${INSTALL_DIR:-/opt/blog-butterfly-go}"
SCRIPT_INVOKED_FROM_PIPE=0
if [[ ! -f "${BASH_SOURCE[0]:-}" || "${BASH_SOURCE[0]:-}" == "bash" || "${BASH_SOURCE[0]:-}" == "-bash" ]]; then
  SCRIPT_INVOKED_FROM_PIPE=1
fi

usage() {
  cat <<'EOF'
用法: ./script/install-docker-compose.sh [选项]

选项:
  --no-build                  不执行 build，直接启动现有镜像/容器
  --down                      停止并删除 compose 服务
  --skip-install-docker       若未安装 docker / docker compose，则直接报错退出，不自动安装
  -h, --help                  显示帮助

环境变量:
  PUBLIC_HOST                 输出访问地址时使用，默认 127.0.0.1
  INSTALL_DIR                 远程安装目录，默认 /opt/blog-butterfly-go
  REPO_URL                    仓库地址，默认 https://github.com/alexclownfish/blog-butterfly-go.git
  REPO_REF                    仓库分支/标签/提交，默认 main

说明:
  - 若系统未安装 Docker，本脚本会自动尝试安装 Docker Engine 与 Docker Compose Plugin
  - 若已安装，则自动跳过，不重复折腾
  - 若通过 curl | bash 远程执行，会先拉取仓库到 INSTALL_DIR，再执行 compose
EOF
}

ensure_runtime_deps() {
  if command_exists curl && command_exists git; then
    return 0
  fi

  detect_os
  require_root_or_sudo

  case "${OS_ID}" in
    ubuntu|debian)
      run_privileged apt-get update
      run_privileged apt-get install -y curl git ca-certificates
      ;;
    centos|rhel|rocky|almalinux|ol|fedora)
      if command_exists dnf; then
        run_privileged dnf -y install curl git ca-certificates
      else
        run_privileged yum -y install curl git ca-certificates
      fi
      ;;
    *)
      case "${OS_ID_LIKE}" in
        *debian*)
          run_privileged apt-get update
          run_privileged apt-get install -y curl git ca-certificates
          ;;
        *rhel*|*fedora*)
          if command_exists dnf; then
            run_privileged dnf -y install curl git ca-certificates
          else
            run_privileged yum -y install curl git ca-certificates
          fi
          ;;
        *)
          echo "❌ 当前系统无法自动安装 curl/git，请先手动安装后再执行。" >&2
          exit 1
          ;;
      esac
      ;;
  esac
}

prepare_repo_if_needed() {
  if [[ -f "$COMPOSE_FILE" && -d "$ROOT_DIR/.git" ]]; then
    return 0
  fi

  if [[ "$SCRIPT_INVOKED_FROM_PIPE" -ne 1 && -f "$COMPOSE_FILE" ]]; then
    return 0
  fi

  log "📥 准备部署仓库到 ${INSTALL_DIR} ..."
  ensure_runtime_deps
  run_privileged mkdir -p "$(dirname "$INSTALL_DIR")"

  if [[ ! -d "$INSTALL_DIR/.git" ]]; then
    if [[ -e "$INSTALL_DIR" ]]; then
      echo "❌ INSTALL_DIR 已存在但不是 git 仓库: $INSTALL_DIR" >&2
      exit 1
    fi
    run_privileged git clone --depth 1 --branch "$REPO_REF" "$REPO_URL" "$INSTALL_DIR"
  else
    run_privileged git -C "$INSTALL_DIR" fetch --depth 1 origin "$REPO_REF"
    run_privileged git -C "$INSTALL_DIR" checkout "$REPO_REF"
    run_privileged git -C "$INSTALL_DIR" reset --hard "origin/$REPO_REF"
  fi

  ROOT_DIR="$INSTALL_DIR"
  COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"

  if [[ ! -f "$COMPOSE_FILE" ]]; then
    echo "❌ 仓库已拉取，但仍未找到 $COMPOSE_FILE" >&2
    exit 1
  fi
}

log() {
  echo "[$(date '+%F %T')] $*"
}

require_root_or_sudo() {
  if [[ "${EUID}" -ne 0 ]] && ! command -v sudo >/dev/null 2>&1; then
    echo "❌ 需要 root 权限或 sudo 命令来安装 Docker" >&2
    exit 1
  fi
}

run_privileged() {
  if [[ "${EUID}" -eq 0 ]]; then
    "$@"
  else
    sudo "$@"
  fi
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

compose_available() {
  docker compose version >/dev/null 2>&1
}

detect_os() {
  if [[ -r /etc/os-release ]]; then
    . /etc/os-release
    OS_ID="${ID:-}"
    OS_VERSION_ID="${VERSION_ID:-}"
    OS_ID_LIKE="${ID_LIKE:-}"
  else
    OS_ID=""
    OS_VERSION_ID=""
    OS_ID_LIKE=""
  fi
}

apt_install_docker() {
  log "📦 检测到 Debian/Ubuntu 系，开始安装 Docker..."
  run_privileged apt-get update
  run_privileged apt-get install -y ca-certificates curl gnupg
  run_privileged install -m 0755 -d /etc/apt/keyrings
  if [[ ! -f /etc/apt/keyrings/docker.gpg ]]; then
    curl -fsSL https://download.docker.com/linux/${OS_ID}/gpg | run_privileged gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    run_privileged chmod a+r /etc/apt/keyrings/docker.gpg
  fi
  local arch
  arch="$(dpkg --print-architecture)"
  echo "deb [arch=${arch} signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/${OS_ID} ${VERSION_CODENAME} stable" | run_privileged tee /etc/apt/sources.list.d/docker.list >/dev/null
  run_privileged apt-get update
  run_privileged apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  run_privileged systemctl enable --now docker
}

yum_install_docker() {
  log "📦 检测到 RHEL/CentOS/Rocky/Alma 系，开始安装 Docker..."
  if command_exists dnf; then
    run_privileged dnf -y install dnf-plugins-core
    run_privileged dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    run_privileged dnf -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  else
    run_privileged yum -y install yum-utils
    run_privileged yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    run_privileged yum -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  fi
  run_privileged systemctl enable --now docker
}

install_docker_if_needed() {
  if command_exists docker && compose_available; then
    log "✅ 已检测到 Docker 与 Docker Compose，跳过安装"
    return 0
  fi

  if [[ "$INSTALL_DOCKER_IF_MISSING" -ne 1 ]]; then
    echo "❌ 未检测到可用的 Docker / Docker Compose，请先安装，或去掉 --skip-install-docker" >&2
    exit 1
  fi

  require_root_or_sudo
  detect_os

  case "${OS_ID}" in
    ubuntu|debian)
      apt_install_docker
      ;;
    centos|rhel|rocky|almalinux|ol|fedora)
      yum_install_docker
      ;;
    *)
      case "${OS_ID_LIKE}" in
        *debian*)
          apt_install_docker
          ;;
        *rhel*|*fedora*)
          yum_install_docker
          ;;
        *)
          echo "❌ 暂不支持自动安装 Docker 的系统: ID=${OS_ID:-unknown}, ID_LIKE=${OS_ID_LIKE:-unknown}" >&2
          echo "请先手动安装 Docker Engine 与 Docker Compose Plugin 后再执行。" >&2
          exit 1
          ;;
      esac
      ;;
  esac

  if ! command_exists docker; then
    echo "❌ Docker 安装后仍不可用，请检查安装日志" >&2
    exit 1
  fi

  if ! compose_available; then
    echo "❌ Docker Compose Plugin 安装后仍不可用，请检查安装日志" >&2
    exit 1
  fi

  log "✅ Docker 与 Docker Compose 安装完成"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-build)
      BUILD_FLAG=""
      ;;
    --down)
      ACTION="down"
      ;;
    --skip-install-docker)
      INSTALL_DOCKER_IF_MISSING=0
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "❌ 未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
  shift
done

prepare_repo_if_needed
install_docker_if_needed

cd "$ROOT_DIR"

if [[ "$ACTION" == "down" ]]; then
  log "🛑 停止并删除 Docker Compose 服务..."
  docker compose -f "$COMPOSE_FILE" down
  exit 0
fi

log "🐳 启动 Docker Compose 服务..."
if [[ -n "$BUILD_FLAG" ]]; then
  docker compose -f "$COMPOSE_FILE" up -d --build
else
  docker compose -f "$COMPOSE_FILE" up -d
fi

cat <<EOF

✅ Docker Compose 部署完成！
前台访问:  http://${PUBLIC_HOST}:8086
后台访问:  http://${PUBLIC_HOST}:8085
后端 API:  http://${PUBLIC_HOST}:8083/api
MySQL:      ${PUBLIC_HOST}:3306

常用命令:
  docker compose -f docker-compose.yml ps
  docker compose -f docker-compose.yml logs -f backend
  docker compose -f docker-compose.yml down
EOF
