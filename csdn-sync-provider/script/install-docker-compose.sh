#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="${BASH_SOURCE[0]:-$0}"
SCRIPT_BASENAME="$(basename "$SCRIPT_PATH")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" 2>/dev/null && pwd || pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." 2>/dev/null && pwd || pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"

REPO_URL="${REPO_URL:-https://github.com/alexclownfish/blog-butterfly-go.git}"
REPO_REF="${REPO_REF:-main}"
INSTALL_DIR="${INSTALL_DIR:-/opt/blog-butterfly-go}"
PROVIDER_SUBDIR="${PROVIDER_SUBDIR:-csdn-sync-provider}"
AUTO_INSTALL_DOCKER="${AUTO_INSTALL_DOCKER:-1}"
SKIP_BUILD=0
DOWN_ONLY=0
FOLLOW_LOGS=0
FORCE_PULL=0
SKIP_INSTALL_DOCKER=0

usage() {
  cat <<'EOF'
Usage: bash script/install-docker-compose.sh [options]

Options:
  --no-build             Start with docker compose up -d (skip --build)
  --down                 Stop and remove the compose stack
  --logs                 Follow csdn-sync-provider logs after startup
  --pull                 Force git fetch/reset during bootstrap
  --skip-install-docker  Do not auto-install Docker when missing
  --install-docker       Force auto-install Docker when missing
  -h, --help             Show this help

Environment overrides:
  REPO_URL               Git repository URL for bootstrap (default: https://github.com/alexclownfish/blog-butterfly-go.git)
  REPO_REF               Git ref/branch to deploy (default: main)
  INSTALL_DIR            Repo checkout dir for bootstrap (default: /opt/blog-butterfly-go)
  PROVIDER_SUBDIR        Provider subdir inside repo (default: csdn-sync-provider)
  AUTO_INSTALL_DOCKER    1 to auto-install Docker when missing, 0 to disable
EOF
}

log() {
  echo "[info] $*"
}

warn() {
  echo "[warn] $*" >&2
}

fail() {
  echo "[error] $*" >&2
  exit 1
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

is_piped_execution() {
  case "$SCRIPT_BASENAME" in
    bash|-bash|sh|-sh) return 0 ;;
  esac
  [[ ! -f "$SCRIPT_PATH" ]]
}

ensure_packages() {
  if command_exists apt-get; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y "$@"
    return
  fi
  if command_exists dnf; then
    dnf install -y "$@"
    return
  fi
  if command_exists yum; then
    yum install -y "$@"
    return
  fi
  fail "unsupported package manager; please install manually: $*"
}

install_docker_if_needed() {
  if command_exists docker && docker compose version >/dev/null 2>&1; then
    return
  fi

  if [[ "$SKIP_INSTALL_DOCKER" -eq 1 || "$AUTO_INSTALL_DOCKER" != "1" ]]; then
    fail "docker / docker compose not available, and auto-install is disabled"
  fi

  log "docker or docker compose missing; attempting automatic installation"
  ensure_packages ca-certificates curl gnupg lsb-release git

  if command_exists apt-get; then
    install -m 0755 -d /etc/apt/keyrings
    if [[ ! -f /etc/apt/keyrings/docker.asc ]]; then
      curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc || \
        curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
      chmod a+r /etc/apt/keyrings/docker.asc
    fi

    local arch codename
    arch="$(dpkg --print-architecture)"
    codename="$(. /etc/os-release && echo "${VERSION_CODENAME:-}")"
    if [[ -z "$codename" ]]; then
      codename="$(lsb_release -cs)"
    fi

    . /etc/os-release
    printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/%s %s stable\n' \
      "$arch" "${ID:-ubuntu}" "$codename" >/etc/apt/sources.list.d/docker.list
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    return
  fi

  if command_exists dnf || command_exists yum; then
    if command_exists dnf; then
      dnf -y install dnf-plugins-core ca-certificates curl git
      dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo || true
      dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    else
      yum install -y yum-utils ca-certificates curl git
      yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo || true
      yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    fi
    systemctl enable --now docker || true
    return
  fi

  fail "automatic Docker installation is not supported on this distro"
}

ensure_bootstrap_prereqs() {
  local missing=()
  command_exists git || missing+=(git)
  command_exists curl || missing+=(curl)
  if [[ ${#missing[@]} -gt 0 ]]; then
    log "installing bootstrap prerequisites: ${missing[*]}"
    ensure_packages ca-certificates "${missing[@]}"
  fi
}

bootstrap_repo_if_needed() {
  local need_bootstrap=0
  if is_piped_execution; then
    need_bootstrap=1
  elif [[ ! -f "$COMPOSE_FILE" ]]; then
    need_bootstrap=1
  fi

  if [[ "$need_bootstrap" -eq 0 && "$FORCE_PULL" -eq 0 ]]; then
    return
  fi

  ensure_bootstrap_prereqs

  if [[ -e "$INSTALL_DIR" && ! -d "$INSTALL_DIR/.git" ]]; then
    fail "INSTALL_DIR exists but is not a git repo: $INSTALL_DIR"
  fi

  if [[ ! -d "$INSTALL_DIR/.git" ]]; then
    log "cloning repo into $INSTALL_DIR"
    mkdir -p "$(dirname "$INSTALL_DIR")"
    git clone --depth 1 --branch "$REPO_REF" "$REPO_URL" "$INSTALL_DIR"
  else
    log "refreshing repo in $INSTALL_DIR"
    git -C "$INSTALL_DIR" remote set-url origin "$REPO_URL"
    git -C "$INSTALL_DIR" fetch --depth 1 origin "$REPO_REF"
    git -C "$INSTALL_DIR" checkout "$REPO_REF" || git -C "$INSTALL_DIR" checkout -B "$REPO_REF" "origin/$REPO_REF"
    git -C "$INSTALL_DIR" reset --hard "origin/$REPO_REF"
    git -C "$INSTALL_DIR" clean -fd
  fi

  ROOT_DIR="$INSTALL_DIR/$PROVIDER_SUBDIR"
  COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"

  [[ -f "$COMPOSE_FILE" ]] || fail "provider compose file not found after bootstrap: $COMPOSE_FILE"
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --no-build)
        SKIP_BUILD=1
        ;;
      --down)
        DOWN_ONLY=1
        ;;
      --logs)
        FOLLOW_LOGS=1
        ;;
      --pull)
        FORCE_PULL=1
        ;;
      --skip-install-docker)
        SKIP_INSTALL_DOCKER=1
        ;;
      --install-docker)
        AUTO_INSTALL_DOCKER=1
        SKIP_INSTALL_DOCKER=0
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        fail "unknown option: $1"
        ;;
    esac
    shift
  done
}

main() {
  parse_args "$@"
  bootstrap_repo_if_needed
  cd "$ROOT_DIR"

  [[ -f "$COMPOSE_FILE" ]] || fail "docker-compose.yml not found: $COMPOSE_FILE"
  [[ -f .env.example ]] || fail ".env.example not found in $ROOT_DIR"

  if [[ ! -f .env ]]; then
    cp .env.example .env
    log ".env missing, created from .env.example"
  fi

  install_docker_if_needed

  docker compose -f "$COMPOSE_FILE" config >/dev/null

  if [[ "$DOWN_ONLY" -eq 1 ]]; then
    docker compose -f "$COMPOSE_FILE" down
    echo "[done] compose stack stopped"
    exit 0
  fi

  if [[ "$SKIP_BUILD" -eq 1 ]]; then
    docker compose -f "$COMPOSE_FILE" up -d
  else
    docker compose -f "$COMPOSE_FILE" up -d --build
  fi

  echo
  echo "[done] csdn-sync-provider is starting"
  echo "- Repo:   $ROOT_DIR"
  echo "- Health: http://127.0.0.1:${CSDN_SYNC_PROVIDER_PORT:-8091}/health"
  echo "- Status: docker compose -f $COMPOSE_FILE ps"
  echo "- Logs:   docker compose -f $COMPOSE_FILE logs -f csdn-sync-provider"
  echo "- Stop:   docker compose -f $COMPOSE_FILE down"

  if [[ "$FOLLOW_LOGS" -eq 1 ]]; then
    docker compose -f "$COMPOSE_FILE" logs -f csdn-sync-provider
  fi
}

main "$@"
