FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct
COPY . .
RUN go mod tidy && CGO_ENABLED=1 GOOS=linux go build -o server .

FROM alpine:3.19

RUN apk add --no-cache libc6-compat

WORKDIR /app

RUN mkdir -p /app/data

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
