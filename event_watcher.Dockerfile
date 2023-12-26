FROM go:1.21-alpine AS builder

COPY . /app
WORKDIR /app
RUN go build -o collector svc/collector

FROM alpine:latest
RUN apk --no-cache add ca-certificates && mkdir /app
WORKDIR /app
COPY --from=builder /app/event_watcher /app/event_watcher
ENTRYPOINT ["/app/event_watcher"]
