FROM golang:1.22.0-alpine3.18

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o myapp

EXPOSE 8080

CMD ["./myapp"]
