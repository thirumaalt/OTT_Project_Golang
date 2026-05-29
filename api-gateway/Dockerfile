FROM golang:1.23 AS builder
#FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
#RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o gateway .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    GOFLAGS=-mod=mod go build -ldflags="-s -w" -o gateway .

FROM scratch
COPY --from=builder /app/gateway /gateway
EXPOSE 8094
ENTRYPOINT ["/gateway"]
