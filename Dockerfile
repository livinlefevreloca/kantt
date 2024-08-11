FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY ./config.toml ./go.mod ./go.sum /app
RUN go mod download
COPY ./svc/ /app/svc/
COPY ./pkg/ /app/pkg/
ENV CGO_ENABLED=1
RUN apk --no-cache add gcc musl-dev

RUN go build -o collector ./svc/collector && \
    chmod +x collector

FROM alpine:latest AS collector
# install sqlite for use in development
RUN apk --no-cache add ca-certificates sqlite && mkdir /app
WORKDIR /app
COPY --from=builder /app/collector /app/collector
COPY --from=builder /app/config.toml /app/config.toml

ENTRYPOINT ["/app/collector"]
