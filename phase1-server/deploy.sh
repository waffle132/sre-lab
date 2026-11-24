#!/usr/bin/env bash
set -e
image="$1"
if [ -z "$image" ]; then
    echo "用法：bash deploy.sh 镜像名称"
    exit 2
fi


docker image inspect "$image" > /dev/null
previous_image=$(docker inspect u-short --format '{{.Image}}')

deploy_and_check(){
    export U_SHORT_IMAGE="$1"
    docker compose up -d --no-deps --pull never \
    --wait --wait-timeout 60 u-short && 
    python3 probe.py --url http://localhost:8080/healthz --timeout 3 &&

    python3 business_probe.py \
        --target-url https://example.com/deploy-check \
        --timeout 3
}
if deploy_and_check "$image"; then
    echo "发布验证通过：$image"
else
    echo "发布失败，尝试恢复镜像：$previous_image"

    if deploy_and_check "$previous_image"; then
       echo "原镜像已恢复，健康与业务检查通过"
    else
        echo "恢复失败，需要人工处理"
    fi
    exit 1
fi    
