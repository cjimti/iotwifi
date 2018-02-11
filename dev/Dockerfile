FROM arm32v6/alpine

RUN apk update
RUN apk add alpine-sdk go bridge hostapd wireless-tools wpa_supplicant dnsmasq iw

RUN mkdir -p /etc/udev/rules.d/ \
 && mkdir -p /etc/hostapd/

COPY ./configs/90-wireless.rules /etc/udev/rules.d/90-wireless.rules
COPY ./configs/dnsmasq.conf /etc/dnsmasq.conf
COPY ./configs/hostapd.conf /etc/hostapd/hostapd.conf

RUN mkdir -p /go/src/github.com/cjimti/iotwifi/

ENV GOPATH /go
WORKDIR /go/src

RUN go get github.com/bhoriuchi/go-bunyan/bunyan
