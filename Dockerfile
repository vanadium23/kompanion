# Dockerfile (Minimal Changes)

# Define target architecture as a build argument.
# BuildKit automatically sets this when using --platform.
ARG TARGETARCH

# Step 1: Modules caching
# Use a multi-arch Go base image. BuildKit will select the correct one based on TARGETARCH.
FROM golang:1.22.5-alpine as modules

# Set GOOS and GOARCH based on the build argument for this stage.
# This ensures go mod download caches for the correct target architecture.
ENV GOOS=linux GOARCH=${TARGETARCH}

COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
# Use a multi-arch Go base image for the builder stage.
FROM golang:1.22.5-alpine as builder
RUN apk add --update gcc musl-dev # Keep update for minimal change

# Set GOOS and GOARCH for the builder environment.
# This will be used by the go build command.
ENV GOOS=linux GOARCH=${TARGETARCH}

# Copy modules cache built for the *same* target architecture
COPY --from=modules /go/pkg /go/pkg

COPY . /app
WORKDIR /app

# Define KOMPANION_VERSION as a build argument
ARG KOMPANION_VERSION=local
ENV KOMPANION_VERSION=$KOMPANION_VERSION

# Build the application. Use the GOOS/GOARCH from the environment.
# REMOVE the hardcoded GOOS=linux GOARCH=amd64.
RUN go build -ldflags "-X main.Version=$KOMPANION_VERSION" -tags migrate -o /bin/app ./cmd/app

# Step 3: Final
# Keep the same base image as the original for minimal change.
# This image is multi-arch, and BuildKit will select the correct variant.
FROM golang:1.22.5-alpine

ENV GIN_MODE=release
WORKDIR /

# Copy web assets from the builder stage
COPY --from=builder /app/web /web

# Copy the architecture-specific binary from the builder stage
COPY --from=builder /bin/app /app

# Copy CA certs as in the original
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the command to run the application
CMD ["/app"]