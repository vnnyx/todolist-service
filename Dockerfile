FROM golang:1.20-alpine as builder
WORKDIR /builder
COPY . .
RUN apk add upx
RUN go mod tidy
RUN go build \
    -ldflags "-s -w" \
    -o main
RUN upx -9 main

FROM alpine:latest
WORKDIR /app
COPY --from=builder /builder/main .
COPY --from=builder /builder/.env .
CMD /app/main server