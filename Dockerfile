FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/intervuz ./cmd/intervuz

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

COPY --from=builder /out/intervuz /usr/local/bin/intervuz
COPY --from=builder /app/GEO_Json /app/GEO_Json
COPY --from=builder /app/image_floor /app/image_floor
COPY --from=builder /app/schedule /app/schedule
COPY --from=builder /app/swagger.yaml /app/swagger.yaml

EXPOSE 8000

ENTRYPOINT ["intervuz"]
