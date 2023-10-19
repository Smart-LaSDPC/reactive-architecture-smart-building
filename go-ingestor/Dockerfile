FROM golang:1.21.3-alpine AS builder

WORKDIR /go/src
COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o /go/bin/app .

FROM alpine:latest

COPY --from=builder /go/bin/app /ingestor
CMD ["/ingestor"]