VERSION=latest
IMAGE_TAG=eu.gcr.io/datio-sistemas/retrievault:$(VERSION)

all: build build-docker

build:
	go build -o docker/retrievault

build-docker:
	cd docker; docker build --label org.label-schema.build-date=$(shell date -u +%Y-%m-%dT%T.$(date -u +%N | cut -c 1-2)Z) -t $(IMAGE_TAG) .; cd -
