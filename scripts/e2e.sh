#!/usr/bin/env bash
set -e
curl -X POST localhost:3333/ingest/manual
sleep 2
curl -s localhost:3333/signals/current | jq -e '.side'
echo "E2E OK"
