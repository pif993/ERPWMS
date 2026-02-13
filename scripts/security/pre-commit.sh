#!/usr/bin/env bash
set -euo pipefail
if command -v gitleaks >/dev/null 2>&1; then gitleaks detect --source .; else echo "gitleaks not installed"; fi
if command -v trufflehog >/dev/null 2>&1; then trufflehog filesystem .; else echo "trufflehog not installed"; fi
