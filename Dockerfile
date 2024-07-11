FROM --platform=linux/amd64 golang:1.22.3 AS builder
ARG TARGETOS TARGETARCH TARGETVARIANT
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH VARIANT=$TARGETVARIANT

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends tar xz-utils

COPY . .

RUN set -eux; \
	case "$GOARCH" in \
		arm) export GOARM="${VARIANT#v}" ;; \
		amd64) export GOAMD64="$VARIANT" ;; \
		*) [ -z "$VARIANT" ] ;; \
	esac; \
	go env | grep -E 'OS=|ARCH=|ARM=|AMD64='; \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o serve main.go

RUN chmod 777 ./minify.sh && ./minify.sh $TARGETOS $TARGETARCH

FROM scratch

WORKDIR /www

COPY ./assets/index.html /www/index.html
COPY ./assets/404.html /404.html
COPY --from=builder /app/serve /serve

EXPOSE 80 443

CMD [ "/serve", "--port", "80" ]
