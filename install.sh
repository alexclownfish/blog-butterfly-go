#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$ROOT_DIR/k8s"
PUBLIC_HOST="${PUBLIC_HOST:-172.28.74.191}"

BUILD_IMAGES=1
DEPLOY_DB=1
USE_EXTERNAL_DB=0

usage() {
  cat <<'EOF'
用法: ./install.sh [选项]

选项:
  --skip-build        跳过 Docker 镜像构建，只执行 Kubernetes 部署
  --skip-db           不部署 k8s/mysql.yaml，仅部署业务服务
  --use-external-db   使用 k8s/mysql-alias.yaml 指向外部 mysql.default.svc.cluster.local
  -h, --help          显示帮助

环境变量:
  PUBLIC_HOST         部署完成后用于输出访问地址，默认 172.28.74.191

说明:
  - 默认行为会构建 backend / web-vue / admin-vue 镜像，并部署内置 MySQL
  - --skip-db 与 --use-external-db 互斥
EOF
}

log() {
  echo "[$(date '+%F %T')] $*"
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "❌ 缺少依赖命令: $cmd" >&2
    exit 1
  fi
}

build_image() {
  local name="$1"
  local dir="$2"
  log "📦 构建 ${name} 镜像..."
  docker build -t "${name}:latest" "$dir"
}

apply_file() {
  local file="$1"
  log "☸️  应用 $(basename "$file")"
  kubectl apply -f "$file"
}

delete_if_exists() {
  local kind="$1"
  local name="$2"
  local namespace="$3"
  if kubectl get "$kind" "$name" -n "$namespace" >/dev/null 2>&1; then
    log "🧹 删除冲突资源: ${kind}/${name}"
    kubectl delete "$kind" "$name" -n "$namespace" --wait=true
  fi
}

wait_rollout() {
  local resource="$1"
  local namespace="$2"
  log "⏳ 等待 ${resource} 就绪..."
  kubectl rollout status "$resource" -n "$namespace" --timeout=300s
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-build)
      BUILD_IMAGES=0
      ;;
    --skip-db)
      DEPLOY_DB=0
      ;;
    --use-external-db)
      USE_EXTERNAL_DB=1
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

if [[ "$DEPLOY_DB" -eq 0 && "$USE_EXTERNAL_DB" -eq 1 ]]; then
  echo "❌ --skip-db 与 --use-external-db 不能同时使用" >&2
  exit 1
fi

require_cmd docker
require_cmd kubectl

log "🚀 开始部署 blog-butterfly-go"

if [[ "$BUILD_IMAGES" -eq 1 ]]; then
  build_image "blog-butterfly-backend" "$ROOT_DIR/backend"
  build_image "blog-butterfly-web-vue" "$ROOT_DIR/web-vue"
  build_image "blog-butterfly-admin-vue" "$ROOT_DIR/admin-vue"
else
  log "⏭️  跳过镜像构建"
fi

apply_file "$K8S_DIR/namespace.yaml"

if [[ "$USE_EXTERNAL_DB" -eq 1 ]]; then
  delete_if_exists deployment mysql blog-butterfly-go
  delete_if_exists service mysql blog-butterfly-go
  delete_if_exists pvc mysql-data blog-butterfly-go
  apply_file "$K8S_DIR/mysql-alias.yaml"
  log "🛢️  当前使用外部 MySQL 别名模式"
elif [[ "$DEPLOY_DB" -eq 1 ]]; then
  delete_if_exists service mysql blog-butterfly-go
  apply_file "$K8S_DIR/mysql.yaml"
  wait_rollout deployment/mysql blog-butterfly-go
  log "🛢️  当前使用命名空间内 MySQL 部署模式"
else
  log "⏭️  跳过数据库部署，保留现有 mysql 服务配置"
fi

apply_file "$K8S_DIR/backend.yaml"
wait_rollout deployment/blog-butterfly-backend blog-butterfly-go

apply_file "$K8S_DIR/web-vue.yaml"
wait_rollout deployment/blog-butterfly-web-vue blog-butterfly-go

apply_file "$K8S_DIR/admin-vue.yaml"
wait_rollout deployment/blog-butterfly-admin-vue blog-butterfly-go

cat <<EOF

✅ 部署完成！
前台访问:  http://${PUBLIC_HOST}:31086
后台访问:  http://${PUBLIC_HOST}:31085
后端 API:  http://${PUBLIC_HOST}:31083/api

建议检查:
  kubectl get pods -n blog-butterfly-go
  kubectl get svc -n blog-butterfly-go
  kubectl logs deployment/blog-butterfly-backend -n blog-butterfly-go --tail=100
EOF