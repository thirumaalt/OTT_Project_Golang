FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o auth-service .

FROM scratch
COPY --from=builder /app/auth-service /auth-service
EXPOSE 8083
ENTRYPOINT ["/auth-service"]
