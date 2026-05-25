FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o analytics-service .

FROM scratch
COPY --from=builder /app/analytics-service /analytics-service
EXPOSE 8086
ENTRYPOINT ["/analytics-service"]
