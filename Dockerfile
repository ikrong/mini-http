FROM --platform=linux/amd64 golang:1.22.3 AS builder
ARG TARGETOS TARGETARCH TARGETVARIANT IS_LOCAL BUILD_TIME
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH VARIANT=$TARGETVARIANT IS_LOCAL=$IS_LOCAL

WORKDIR /app

COPY . .

RUN set -eux; \
    case "$GOARCH" in \
        arm) export GOARM="${VARIANT#v}" ;; \
        amd64) export GOAMD64="$VARIANT" ;; \
        *) [ -z "$VARIANT" ] ;; \
    esac; \
    go env | grep -E 'OS=|ARCH=|ARM=|AMD64='; \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o serve main.go

RUN ./minify.sh

FROM scratch

LABEL maintainer="ikrong <contact@ikrong.com>" \
      org.opencontainers.image.authors="ikrong <contact@ikrong.com>" \
      org.opencontainers.image.source="https://github.com/ikrong/mini-http" \
      org.opencontainers.image.description="A minimal static server" \
      org.opencontainers.image.license="MIT"

WORKDIR /www

COPY ./assets/index.html /www/index.html
COPY ./assets/404.html /404.html
COPY --from=builder /app/serve /usr/bin/serve

EXPOSE 80 443

ENTRYPOINT [ "serve" ]

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
        CMD [ "serve", "get", "http://localhost" ]

CMD [ "--port", "80" ]
