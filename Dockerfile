FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main /main

EXPOSE 8080

CMD ["/main"]