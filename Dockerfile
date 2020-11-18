FROM golang:1.15.4-alpine

WORKDIR /src

COPY . .

RUN go build -o /bin/cfc_heartbeat .

CMD ["/bin/cfc_heartbeat"]
