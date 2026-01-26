#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TERRAFORM_DIR="$ROOT_DIR/terraform"

terraform -chdir="$TERRAFORM_DIR" apply -var enable_cloudsql=false
terraform -chdir="$TERRAFORM_DIR" destroy
