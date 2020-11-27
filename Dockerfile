FROM golang:1.15.4-alpine

WORKDIR /src
COPY . .

RUN go build -o /bin/heartbeat_server .

CMD ["/bin/heartbeat_server"]
