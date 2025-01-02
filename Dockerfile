FROM golang:1.23 AS builder

WORKDIR /app
COPY . .
RUN go get -d -v
RUN go build .

FROM ubuntu:24.04

COPY --from=builder /app/ovn-exporter /ovn-exporter

RUN apt-get update && apt-get install -y \
    ovn-common \
    openvswitch-common \
    && rm -rf /var/lib/apt/lists/*

ENTRYPOINT [ "/ovn-exporter" ]
