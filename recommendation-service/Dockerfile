FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o recommendation-service .

FROM scratch
COPY --from=builder /app/recommendation-service /recommendation-service
EXPOSE 8087
ENTRYPOINT ["/recommendation-service"]
