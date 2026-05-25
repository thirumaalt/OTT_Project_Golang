FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o payment-service .

FROM scratch
COPY --from=builder /app/payment-service /payment-service
EXPOSE 8091
ENTRYPOINT ["/payment-service"]
