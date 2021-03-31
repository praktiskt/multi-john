
make-random:
	uuidgen | tail -c 2 | tr -d '\n' | sha256sum | sed 's/[^a-z0-9]//g' | tr -d '\n' > dummy

run:  #make-random
	go run *.go

standalone-etcd:
	docker run \
	-e ALLOW_NONE_AUTHENTICATION=yes \
    -e ETCD_ADVERTISE_CLIENT_URLS=http://localhost:2379 \
	-p 2379:2379 \
	-p 2380:2380 \
	bitnami/etcd:latest
