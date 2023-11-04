FROM golang:1.21.1

RUN mkdir -p /app

WORKDIR /app

ADD . /app

RUN go build ./p2c.go

CMD ["./p2c"]
