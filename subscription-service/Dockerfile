FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o subscription-service .

FROM scratch
COPY --from=builder /app/subscription-service /subscription-service
EXPOSE 8093
ENTRYPOINT ["/subscription-service"]
