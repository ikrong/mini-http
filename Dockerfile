FROM golang:alpine3.17 AS Builder

WORKDIR /app

COPY . /app

RUN go build -ldflags="-s -w" -o serve main.go

RUN ./libs/upx-amd64 --best ./serve

FROM scratch

COPY --from=Builder /app/serve /serve

COPY --from=Builder /app/assets/index.html /www/index.html

COPY --from=Builder /app/assets/404.html /404.html

WORKDIR /www

CMD [ "/serve", "--port", "80" ]