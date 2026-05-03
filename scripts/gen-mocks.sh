#!/usr/bin/env bash
set -euo pipefail

echo "Generating mocks..."
go run github.com/vektra/mockery/v2
echo "Done."
