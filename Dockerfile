# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o wa-gateway-go .

# Stage 2: Runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/wa-gateway-go .
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser

EXPOSE 4010

ENTRYPOINT ["./wa-gateway-go"]
