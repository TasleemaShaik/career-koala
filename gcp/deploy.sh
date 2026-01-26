#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ "${SKIP_HELM_DEP_UPDATE:-}" != "1" ]]; then
  helm dependency update "$ROOT_DIR/helm/career-koala"
fi

: "${GCR:?GCR is required}"
: "${GCR_PROJECT:?GCR_PROJECT is required}"
: "${GCR_IMAGE_API:?GCR_IMAGE_API is required}"
: "${GCR_IMAGE_UI:?GCR_IMAGE_UI is required}"
: "${GCR_IMAGE_MIGRATE:?GCR_IMAGE_MIGRATE is required}"
: "${GCR_IMAGE_TAG:?GCR_IMAGE_TAG is required}"
: "${API_BASE:?API_BASE is required}"
: "${INGRESS_HOST:?INGRESS_HOST is required}"
: "${GOOGLE_CLOUD_PROJECT:?GOOGLE_CLOUD_PROJECT is required}"
: "${MODEL_NAME:?MODEL_NAME is required}"
: "${VERTEX_LOCATION:?VERTEX_LOCATION is required}"
: "${POSTGRES_HOST:?POSTGRES_HOST is required}"
: "${POSTGRES_PORT:?POSTGRES_PORT is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${POSTGRES_SSLMODE:?POSTGRES_SSLMODE is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"

API_REPO="${GCR}/${GCR_PROJECT}/${GCR_IMAGE_API}"
UI_REPO="${GCR}/${GCR_PROJECT}/${GCR_IMAGE_UI}"
MIGRATE_REPO="${GCR}/${GCR_PROJECT}/${GCR_IMAGE_MIGRATE}"

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx >/dev/null 2>&1 || true
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx --create-namespace

HELM_ARGS=(
  -n career-koala
  --create-namespace
  -f "$ROOT_DIR/gcp/values.yaml"
  --set-string api.image.repository="$API_REPO"
  --set-string api.image.tag="$GCR_IMAGE_TAG"
  --set-string ui.image.repository="$UI_REPO"
  --set-string ui.image.tag="$GCR_IMAGE_TAG"
  --set-string migrations.image.repository="$MIGRATE_REPO"
  --set-string migrations.image.tag="$GCR_IMAGE_TAG"
  --set-string ui.env.API_BASE="$API_BASE"
  --set-string ingress.host="$INGRESS_HOST"
  --set-string api.env.GOOGLE_CLOUD_PROJECT="$GOOGLE_CLOUD_PROJECT"
  --set-string api.env.MODEL_NAME="$MODEL_NAME"
  --set-string api.env.VERTEX_LOCATION="$VERTEX_LOCATION"
  --set-string api.env.POSTGRES_HOST="$POSTGRES_HOST"
  --set-string api.env.POSTGRES_PORT="$POSTGRES_PORT"
  --set-string api.env.POSTGRES_USER="$POSTGRES_USER"
  --set-string api.env.POSTGRES_DB="$POSTGRES_DB"
  --set-string api.env.POSTGRES_SSLMODE="$POSTGRES_SSLMODE"
  --set-string api.secret.POSTGRES_PASSWORD="$POSTGRES_PASSWORD"
)

# printf 'HELM_ARGS=%q\n' "${HELM_ARGS[@]}"
helm upgrade --install career-koala "$ROOT_DIR/helm/career-koala" "${HELM_ARGS[@]}"
