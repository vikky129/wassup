FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o cmd/main ./cmd

WORKDIR /app/cmd

EXPOSE 8080
CMD ["./main"]
