IMAGE    ?= cjimti/iotwifi:armhf-1.0.1
NAME     ?= iotwifi

all: build push

dev: dev_build dev_run

build:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)

dev_build:
	docker build -t $(IMAGE) ./dev/

dev_run:
	sudo docker run --rm -it --privileged --network=host \
                   -v $(CURDIR):/go/src/github.com/cjimti/iotwifi \
                   -w /go/src/github.com/cjimti/iotwifi \
                   --name=$(NAME) $(IMAGE)
