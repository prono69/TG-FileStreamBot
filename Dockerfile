FROM --platform=$BUILDPLATFORM golang:1.25-alpine3.21 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app

# Copy everything
COPY . .

# Force Go to reconcile dependencies and update go.sum inside the container
RUN go mod tidy

# Now build with the updated state
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /app/fsb -ldflags="-w -s" ./cmd/fsb

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/fsb /app/fsb
EXPOSE ${PORT}
ENTRYPOINT ["/app/fsb", "run"]
