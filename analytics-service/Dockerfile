FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
#RUN GOFLAGS=-mod=mod go build -ldflags="-s -w" -o analytics-service .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    GOFLAGS=-mod=mod go build -ldflags="-s -w" -o analytics-service .

FROM scratch
COPY --from=builder /app/analytics-service /analytics-service
EXPOSE 8086
ENTRYPOINT ["/analytics-service"]
