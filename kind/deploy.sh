#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

helm dependency update "$ROOT_DIR/helm/career-koala"

: "${CR:?CR is required}"
: "${CR_PROJECT:?CR_PROJECT is required}"
: "${CR_IMAGE_API:?CR_IMAGE_API is required}"
: "${CR_IMAGE_UI:?CR_IMAGE_UI is required}"
: "${CR_IMAGE_MIGRATE:?CR_IMAGE_MIGRATE is required}"
: "${CR_IMAGE_TAG:?CR_IMAGE_TAG is required}"
: "${API_BASE:?API_BASE is required}"
: "${GOOGLE_CLOUD_PROJECT:?GOOGLE_CLOUD_PROJECT is required}"
: "${MODEL_NAME:?MODEL_NAME is required}"
: "${VERTEX_LOCATION:?VERTEX_LOCATION is required}"
: "${POSTGRES_HOST:?POSTGRES_HOST is required}"
: "${POSTGRES_PORT:?POSTGRES_PORT is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${POSTGRES_SSLMODE:?POSTGRES_SSLMODE is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"

API_REPO="${CR}/${CR_PROJECT}/${CR_IMAGE_API}"
UI_REPO="${CR}/${CR_PROJECT}/${CR_IMAGE_UI}"
MIGRATE_REPO="${CR}/${CR_PROJECT}/${CR_IMAGE_MIGRATE}"

HELM_ARGS=(
  -n career-koala
  --create-namespace
  -f "$ROOT_DIR/kind/values.yaml"
  --set-string api.image.repository="$API_REPO"
  --set-string api.image.tag="$CR_IMAGE_TAG"
  --set-string ui.image.repository="$UI_REPO"
  --set-string ui.image.tag="$CR_IMAGE_TAG"
  --set-string migrations.image.repository="$MIGRATE_REPO"
  --set-string migrations.image.tag="$CR_IMAGE_TAG"
  --set-string ui.env.API_BASE="$API_BASE"
  --set-string api.env.GOOGLE_CLOUD_PROJECT="$GOOGLE_CLOUD_PROJECT"
  --set-string api.env.MODEL_NAME="$MODEL_NAME"
  --set-string api.env.VERTEX_LOCATION="$VERTEX_LOCATION"
  --set-string api.env.POSTGRES_HOST="$POSTGRES_HOST"
  --set-string api.env.POSTGRES_PORT="$POSTGRES_PORT"
  --set-string api.env.POSTGRES_USER="$POSTGRES_USER"
  --set-string api.env.POSTGRES_DB="$POSTGRES_DB"
  --set-string api.env.POSTGRES_SSLMODE="$POSTGRES_SSLMODE"
  --set-string api.secret.POSTGRES_PASSWORD="$POSTGRES_PASSWORD"
  --set-string postgresql.auth.username="$POSTGRES_USER"
  --set-string postgresql.auth.password="$POSTGRES_PASSWORD"
  --set-string postgresql.auth.database="$POSTGRES_DB"
)

helm upgrade --install career-koala "$ROOT_DIR/helm/career-koala" "${HELM_ARGS[@]}"
