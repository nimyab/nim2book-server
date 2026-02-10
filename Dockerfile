FROM golang:1.25-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go mod verify

RUN go build -o main ./cmd/main.go

EXPOSE 5050

CMD ["./main"]