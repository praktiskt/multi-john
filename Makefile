CURRENT_COMMIT=$(shell git rev-parse HEAD)
IMG_NAME=multi-john

run: 
	go run *.go

standalone-etcd:
	docker run \
	-e ALLOW_NONE_AUTHENTICATION=yes \
    -e ETCD_ADVERTISE_CLIENT_URLS=http://localhost:2379 \
	-p 2379:2379 \
	-p 2380:2380 \
	bitnami/etcd:latest

bup:
	docker-compose build && docker-compose up

build:
	docker build . -t multi-john:latest

release: build
	docker tag multi-john:latest praktiskt/${IMG_NAME}:latest &&\
	docker tag multi-john:latest praktiskt/${IMG_NAME}:${CURRENT_COMMIT} &&\
	docker push praktiskt/${IMG_NAME}:latest &&\
	docker push praktiskt/${IMG_NAME}:${CURRENT_COMMIT}
