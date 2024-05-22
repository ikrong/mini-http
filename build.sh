#!/usr/bin/env bash

OS=$1
OSArch=$(dpkg --print-architecture)

case "$OSArch" in
    ppc64el)      OSArch="powerpc64le" ;;
    armhf)      OSArch="arm" ;;
esac

if [ "$OS" = "darwin" ]; then
    exit 0
fi

UPXVersion=$(wget -qO- -t1 -T2 "https://github.com/upx/upx/releases" | grep -oP 'UPX \K([0-9]+\.[0-9]+\.[0-9]+)' | head -n 1)

upxFileName="upx-${UPXVersion}-${OSArch}_${OS}"
upxDownloadUrl="https://github.com/upx/upx/releases/download/v${UPXVersion}/${upxFileName}.tar.xz"

if [ ! -f "./upx/$upxFileName/upx" ]; then
    mkdir -p ./upx
    wget $upxDownloadUrl -O ./upx/upx.tar.xz
    tar -xvf ./upx/upx.tar.xz -C ./upx
fi

./upx/$upxFileName/upx --best /app/serve