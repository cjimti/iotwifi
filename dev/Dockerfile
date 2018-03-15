FROM arm32v6/alpine

RUN apk update
RUN apk add alpine-sdk go bridge hostapd wireless-tools wireless-tools-dev wpa_supplicant dnsmasq iw

RUN mkdir -p /etc/wpa_supplicant/
COPY ./configs/wpa_supplicant.conf /etc/wpa_supplicant/wpa_supplicant.conf

RUN mkdir -p /go/src/github.com/cjimti/iotwifi/

ENV GOPATH /go
WORKDIR /go/src

RUN go get github.com/bhoriuchi/go-bunyan/bunyan
RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/handlers
