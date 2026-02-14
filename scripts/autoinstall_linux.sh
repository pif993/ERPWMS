#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/pif993/ERPWMS.git}"
INSTALL_DIR="${INSTALL_DIR:-/opt/erpwms}"

ADMIN_EMAIL="${ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-StrongPassw0rd!}"
AUTOTEST_TOKEN="${AUTOTEST_TOKEN:-}"

need_root() {
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "Run as root (sudo)."
    exit 1
  fi
}

detect_os() {
  if [[ -f /etc/os-release ]]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    echo "${ID:-unknown}"
  else
    echo "unknown"
  fi
}

install_docker_debian() {
  apt-get update -y
  apt-get install -y ca-certificates curl gnupg lsb-release git openssl
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
    > /etc/apt/sources.list.d/docker.list
  apt-get update -y
  apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
}

install_docker_rhel() {
  dnf install -y dnf-plugins-core git openssl curl
  dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
  dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
}

gen_secret_b64() { openssl rand -base64 32; }

write_env() {
  local env_file="$1"
  local jwt_cur jwt_prev search audit field autotok
  jwt_cur="$(gen_secret_b64)"
  jwt_prev="$(gen_secret_b64)"
  search="$(gen_secret_b64)"
  audit="$(gen_secret_b64)"
  field="$(gen_secret_b64)"
  autotok="${AUTOTEST_TOKEN:-}"
  [[ -n "${autotok}" ]] || autotok="$(gen_secret_b64)"

  cat > "$env_file" <<EOF_ENV
ENV=dev
HTTP_ADDR=:8080
COOKIE_SECURE=false
CORS_ALLOWED_ORIGINS=http://localhost:8080
RATE_LIMIT_LOGIN_PER_MIN=10
RATE_LIMIT_API_PER_MIN=120

POSTGRES_DB=erpwms
POSTGRES_SUPER_USER=postgres
POSTGRES_SUPER_PASSWORD=postgres
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

APP_DB_USER=erp_app
APP_DB_PASSWORD=change-me-app
DB_URL=postgres://erp_app:change-me-app@postgres:5432/erpwms?sslmode=disable

REDIS_ADDR=redis:6379
NATS_URL=nats://nats:4222

JWT_ISSUER=erpwms
JWT_AUDIENCE=erpwms-users
JWT_SIGNING_KEY_CURRENT=${jwt_cur}
JWT_SIGNING_KEY_PREVIOUS=${jwt_prev}

SEARCH_PEPPER=${search}
AUDIT_PEPPER=${audit}

FIELD_ENC_MASTER_KEY_CURRENT=${field}
FIELD_ENC_MASTER_KEY_PREVIOUS=
FIELD_ENC_KEY_ID_CURRENT=v1
FIELD_ENC_KEY_ID_PREVIOUS=v0

ADMIN_EMAIL=${ADMIN_EMAIL}
ADMIN_PASSWORD=${ADMIN_PASSWORD}

AUTOTEST_ENABLED=true
AUTOTEST_TOKEN=${autotok}

ANALYTICS_SERVICE_TOKEN=change-analytics-token
EOF_ENV
}

main() {
  need_root

  case "$(detect_os)" in
    ubuntu|debian) install_docker_debian ;;
    rhel|centos|rocky|almalinux|fedora) install_docker_rhel ;;
    *) echo "Unsupported OS"; exit 1 ;;
  esac

  mkdir -p "$INSTALL_DIR"
  if [[ ! -d "$INSTALL_DIR/.git" ]]; then
    git clone "$REPO_URL" "$INSTALL_DIR"
  fi
  cd "$INSTALL_DIR"

  [[ -f infra/.env ]] || write_env infra/.env

  docker compose -f infra/docker-compose.yml up -d --build
  docker compose -f infra/docker-compose.yml run --rm migrator
  docker compose -f infra/docker-compose.yml run --rm seeder

  echo "Install complete."
  echo "Portal (Caddy): http://<server-ip>:8080"
  echo "API health:     http://<server-ip>:8081/health"
  echo "Autotest GUI:   http://<server-ip>:8081/autotest"
}

main "$@"
