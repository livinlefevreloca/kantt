FROM golang:1.21-alpine AS builder

COPY . /app
WORKDIR /app
RUN go build -o collector ./svc/collector && \
    chmod +x collector

FROM alpine:latest AS collector
RUN apk --no-cache add ca-certificates && mkdir /app
WORKDIR /app
COPY --from=builder /app/collector /app/collector
ENTRYPOINT ["/app/collector"]
