#!/usr/bin/env bash

set -e

OS=$TARGETOS
OSArch=$(dpkg --print-architecture)
ORIGIN="https://github.com"

if [ "$IS_LOCAL" != "" ]; then
    echo "Building in local"
    # 修改镜像源为阿里云
    sed -i 's|deb.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list.d/debian.sources
fi

case "$OSArch" in
    ppc64el)      OSArch="powerpc64le" ;;
    armhf)        OSArch="arm" ;;
esac

if [ "$OS" = "darwin" ]; then
    exit 0
fi

apt-get update
apt-get install -y --no-install-recommends tar xz-utils

UPXVersion=$(wget -qO- -t1 -T2 "$ORIGIN/upx/upx/releases" | grep -oP 'UPX \K([0-9]+\.[0-9]+\.[0-9]+)' | head -n 1)

upxFileName="upx-${UPXVersion}-${OSArch}_${OS}"
upxDownloadUrl="$ORIGIN/upx/upx/releases/download/v${UPXVersion}/${upxFileName}.tar.xz"

if [ ! -f "./upx/$upxFileName/upx" ]; then
    mkdir -p ./upx
    wget $upxDownloadUrl -O ./upx/upx.tar.xz
    tar -xvf ./upx/upx.tar.xz -C ./upx
fi

./upx/$upxFileName/upx --best /app/serve