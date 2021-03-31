package main

import (
	"log"
	"os"
	"time"

	"github.com/magnusfurugard/multi-john/node"
	"go.etcd.io/etcd/clientv3"
)

func main() {

	endpoint := []string{}
	if len(os.Getenv("ETCD_ADVERTISE_CLIENT_URLS")) != 0 {
		endpoint = append(endpoint, os.Getenv("ETCD_ADVERTISE_CLIENT_URLS"))
	} else {
		endpoint = append(endpoint, "localhost:2379")
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		panic(err)
	}
	defer cli.Close()

	for i := 0; i < 10; i++ {
		_, err := node.New(5, cli)
		if err != nil {
			log.Print(err)
		}
	}
}
