# ----------------- Builder Image
FROM golang:1.15-alpine3.12 as builder

WORKDIR /app
RUN set -eux; \
  apk add --no-cache gcc git libc-dev; \
  go get -tags musl -u golang.org/x/lint/golint;

COPY go.mod go.sum ./
RUN go mod tidy

COPY pkg ./pkg/

RUN set -eux; \
  ${GOPATH}/bin/golint ./...; \
  go test -tags musl ./...; \
  go build -tags musl -o ./ ./...

# ----------------- Runtime Image
FROM alpine:3.12

WORKDIR /app

# Copy our executable from the builder image
COPY --from=builder /app/driver ./driver

#EXPOSE 8080/tcp 9203/tcp

CMD ["./driver"]
