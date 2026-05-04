# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o /out/app .

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S nonroot && adduser -S nonroot -G nonroot

WORKDIR /

COPY --from=builder /out/app /app
COPY migrations/ /migrations/

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app"]
CMD ["serve"]
