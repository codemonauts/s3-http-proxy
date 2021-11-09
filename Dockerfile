FROM golang:1.17 AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o proxy .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /src/proxy .
CMD ["./proxy"]
