# syntax=docker/dockerfile:1

FROM golang:1.21 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/datagen ./cmd/datagen
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ingest ./cmd/ingest

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /out/datagen /app/datagen
COPY --from=builder /out/ingest /app/ingest

ENV SERVER_PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
