FROM golang:1.22.0-alpine3.18

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build  -a -installsuffix cgo -o myapp
EXPOSE 8080

CMD ["./myapp"]
