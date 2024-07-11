#!/bin/bash

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Please input version"
    exit 1
fi

PLATFORM="linux/amd64"

if [ "$(uname)" = "Darwin" ]; then
    PLATFORM="linux/arm64"
fi

docker buildx build \
       --progress plain \
       --platform $PLATFORM \
       -f Dockerfile \
       -t ikrong/mini-http:$VERSION \
       --load \
       . 
