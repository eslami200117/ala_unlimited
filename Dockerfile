FROM golang:1.24-alpine3.22 AS builder

WORKDIR /app

RUN apk add --no-cache protobuf

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

ENV PATH="/go/bin:${PATH}"

COPY . .

RUN go mod tidy
RUN go mod download

RUN protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  protocpb/stream.proto

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app


FROM alpine:3.22 AS build-release-stage

WORKDIR /

COPY --from=builder /bin/app /bin/app

COPY --from=builder /app/admins.txt /bin/admins.txt
COPY --from=builder /app/.env /bin/.env

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot:nonroot

WORKDIR /bin

ENTRYPOINT ["./app"]
