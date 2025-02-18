# Step 1: Modules caching
FROM golang:1.22.5-alpine as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.22.5-alpine as builder
RUN apk add --update gcc musl-dev
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app

# Step 3: Final
FROM golang:1.22.5-alpine
WORKDIR /
COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/web /web
COPY --from=builder /bin/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]
