FROM arm32v6/alpine

RUN apk update
RUN apk add alpine-sdk go bridge hostapd wireless-tools wireless-tools-dev wpa_supplicant dnsmasq iw

RUN mkdir -p /etc/network/
RUN mkdir -p /etc/wpa_supplicant/
RUN mkdir -p /etc/hostapd/

COPY ./configs/wpa_supplicant.conf /etc/wpa_supplicant/wpa_supplicant.conf
COPY ./configs/hostapd.conf /etc/hostapd/hostapd.conf
COPY ./configs/interfaces /etc/network/interfaces

RUN mkdir -p /go/src/github.com/cjimti/iotwifi/

ENV GOPATH /go
WORKDIR /go/src

RUN go get github.com/bhoriuchi/go-bunyan/bunyan

