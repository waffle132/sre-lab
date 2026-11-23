#!/usr/bin/env bash
set -e
image="$1"
if [ -z "$image" ]; then
    echo "用法：bash deploy.sh 镜像名称"
    exit 2
fi

export U_SHORT_IMAGE="$image"

docker image inspect "$image" > /dev/null

docker compose up -d --no-deps \
    --wait --wait-timeout 60 u-short

python3 probe.py --url http://localhost:8080/healthz --timeout 3

python3 business_probe.py \
    --target-url https://example.com/deploy-check \
    --timeout 3

echo "发布验证通过：$image"