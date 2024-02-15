# stage de build
FROM golang:1.22.0 AS build

WORKDIR /app

COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o rinha

FROM scratch

WORKDIR /app

COPY --from=build /app/rinha ./

EXPOSE 8080

CMD [ "./rinha" ]