FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
#RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o watchhistory-service .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    GOFLAGS=-mod=mod go build -ldflags="-s -w" -o watchhistory-service .

FROM scratch
COPY --from=builder /app/watchhistory-service /watchhistory-service
EXPOSE 8090
ENTRYPOINT ["/watchhistory-service"]
