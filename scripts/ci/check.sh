#!/usr/bin/env bash
set -euo pipefail
make lint-go test-go sec-go lint-py test-py sec-py
