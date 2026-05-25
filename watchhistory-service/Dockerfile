FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o watchhistory-service .

FROM scratch
COPY --from=builder /app/watchhistory-service /watchhistory-service
EXPOSE 8090
ENTRYPOINT ["/watchhistory-service"]
