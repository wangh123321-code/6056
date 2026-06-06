FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server .

FROM alpine:3.19

RUN apk add --no-cache libc6-compat

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/workshop.db* ./

EXPOSE 8080

CMD ["./server"]
