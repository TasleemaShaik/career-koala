#!/usr/bin/env python3
import argparse
import json
import re
import subprocess
import sys
from pathlib import Path
from urllib.parse import urlparse, parse_qs, unquote


def run_terraform_output(terraform_dir: Path):
    try:
        result = subprocess.run(
            ["terraform", "output", "-json"],
            cwd=terraform_dir,
            check=True,
            capture_output=True,
            text=True,
        )
    except FileNotFoundError:
        raise RuntimeError("terraform binary not found in PATH")
    except subprocess.CalledProcessError as exc:
        raise RuntimeError(exc.stderr.strip() or "terraform output failed")
    return json.loads(result.stdout)


def get_output(outputs, key):
    entry = outputs.get(key)
    if entry is None:
        return ""
    if isinstance(entry, dict):
        return entry.get("value") or ""
    return entry or ""


def parse_database_url(url):
    if not url:
        return {}
    parsed = urlparse(url)
    if parsed.scheme not in ("postgres", "postgresql"):
        return {}
    user = unquote(parsed.username or "")
    password = unquote(parsed.password or "")
    host = parsed.hostname or ""
    port = str(parsed.port or "") if parsed.port else ""
    db = parsed.path.lstrip("/") if parsed.path else ""
    query = parse_qs(parsed.query or "")
    sslmode = query.get("sslmode", [""])[0]
    return {
        "POSTGRES_USER": user,
        "POSTGRES_PASSWORD": password,
        "POSTGRES_HOST": host,
        "POSTGRES_PORT": port,
        "POSTGRES_DB": db,
        "POSTGRES_SSLMODE": sslmode,
    }


def parse_tfvars(tfvars_path: Path):
    if not tfvars_path.exists():
        return {}
    data = tfvars_path.read_text(encoding="utf-8")
    updates = {}
    match = re.search(r'^\s*project_id\s*=\s*"([^"]+)"\s*$', data, re.M)
    if match:
        updates["GOOGLE_CLOUD_PROJECT"] = match.group(1)
    match = re.search(r'^\s*region\s*=\s*"([^"]+)"\s*$', data, re.M)
    if match:
        updates["VERTEX_LOCATION"] = match.group(1)
    return updates


def parse_plain_output(text):
    outputs = {}
    for line in text.splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        match = re.match(r"^([A-Za-z0-9_]+)\s*=\s*(.+)$", line)
        if not match:
            continue
        key, raw_value = match.groups()
        value = raw_value.strip()
        if value.startswith('"') and value.endswith('"'):
            value = value[1:-1]
        if value == "<sensitive>":
            value = ""
        outputs[key] = {"value": value}
    return outputs


def yaml_quote(value):
    escaped = value.replace("\\", "\\\\").replace('"', '\\"')
    return f"\"{escaped}\""


def update_values_file(values_path: Path, updates):
    lines = values_path.read_text(encoding="utf-8").splitlines()
    missing = []
    for key, value in updates.items():
        if not value:
            continue
        pattern = re.compile(rf"^(\s*{re.escape(key)}:\s*).*$")
        replaced = False
        new_lines = []
        for line in lines:
            if pattern.match(line):
                new_lines.append(pattern.sub(rf"\1{yaml_quote(value)}", line))
                replaced = True
            else:
                new_lines.append(line)
        lines = new_lines
        if not replaced:
            missing.append(key)
    values_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return missing


def main():
    parser = argparse.ArgumentParser(description="Update gcp/values.yaml from terraform outputs.")
    repo_root = Path(__file__).resolve().parents[1]
    parser.add_argument(
        "--terraform-dir",
        default=str(repo_root / "terraform"),
        help="Path to the terraform directory (default: ../terraform)",
    )
    parser.add_argument(
        "--values-file",
        default=str(repo_root / "gcp" / "values.yaml"),
        help="Path to values.yaml (default: gcp/values.yaml)",
    )
    parser.add_argument(
        "--outputs-json",
        default="",
        help="Optional path to terraform output -json file",
    )
    parser.add_argument(
        "--outputs-text",
        default="",
        help="Optional path to terraform output text file (key = value format)",
    )
    parser.add_argument(
        "--tfvars",
        default=str(repo_root / "terraform" / "terraform.tfvars"),
        help="Path to terraform.tfvars (default: ../terraform/terraform.tfvars)",
    )
    parser.add_argument(
        "--api-base",
        default="",
        help="Override API_BASE value in values.yaml",
    )
    args = parser.parse_args()

    terraform_dir = Path(args.terraform_dir)
    values_path = Path(args.values_file)
    outputs = {}

    if args.outputs_text:
        outputs = parse_plain_output(Path(args.outputs_text).read_text(encoding="utf-8"))
    elif args.outputs_json:
        outputs = json.loads(Path(args.outputs_json).read_text(encoding="utf-8"))
    else:
        outputs = run_terraform_output(terraform_dir)

    updates = {}
    db_url = get_output(outputs, "database_url")
    updates.update(parse_database_url(db_url))

    if not updates.get("POSTGRES_HOST"):
        updates["POSTGRES_HOST"] = get_output(outputs, "cloudsql_private_ip")
    if not updates.get("POSTGRES_PASSWORD"):
        updates["POSTGRES_PASSWORD"] = get_output(outputs, "cloudsql_password")
    if not updates.get("POSTGRES_DB"):
        updates["POSTGRES_DB"] = get_output(outputs, "cloudsql_database")

    updates.update(parse_tfvars(Path(args.tfvars)))

    if args.api_base:
        updates["API_BASE"] = args.api_base

    missing = update_values_file(values_path, updates)

    print(f"Updated {values_path}")
    for key in sorted(k for k, v in updates.items() if v):
        print(f"- {key}")
    if missing:
        print("Warning: keys not found in values.yaml:", ", ".join(missing))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"error: {exc}", file=sys.stderr)
        sys.exit(1)
