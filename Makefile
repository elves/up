%: %.go
	go build -o $@ $<

docker: up
	docker build

all: docker

.PHONY: docker all
