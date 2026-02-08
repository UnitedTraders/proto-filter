FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /proto-filter .

FROM alpine:3.21
COPY --from=builder /proto-filter /usr/local/bin/proto-filter
ENTRYPOINT ["proto-filter"]
