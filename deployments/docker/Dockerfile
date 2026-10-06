FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o sigif-server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/sigif-server .
COPY --from=builder /app/configs ./configs

EXPOSE 4600

ENTRYPOINT ["./sigif-server"]