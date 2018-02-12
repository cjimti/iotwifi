IMAGE    ?= iotwifi
NAME     ?= iotwifi

dev: dev_build dev_run

dev_build:
	docker build -t $(IMAGE) ./dev/

dev_run:
	sudo docker run --rm -it --privileged --network=host \
                   -v $(CURDIR):/go/src/github.com/cjimti/iotwifi \
                   -w /go/src/github.com/cjimti/iotwifi \
                   --name=$(NAME) $(IMAGE)
