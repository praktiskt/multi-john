package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/magnusfurugard/multi-john/john"
	"github.com/magnusfurugard/multi-john/node"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

func main() {

	endpoint := []string{}
	if len(os.Getenv("ETCD_ADVERTISE_CLIENT_URLS")) != 0 {
		endpoint = append(endpoint, os.Getenv("ETCD_ADVERTISE_CLIENT_URLS"))
	} else {
		endpoint = append(endpoint, "localhost:2379")
	}

	log.Print("connect to etcd...")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	defer cli.Close()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	var totalNodes int
	if len(os.Getenv("TOTAL_NODES")) != 0 {
		totalNodes, _ = strconv.Atoi(os.Getenv("TOTAL_NODES"))
	} else {
		totalNodes = 2
	}

	n, err := node.New(totalNodes, cli, logger)
	if err != nil {
		sugar.Errorf("Unable to start node: %v", err)
		cli.Close()
		os.Exit(1)
	}

	flags := map[string]string{}
	flags["format"] = "raw-sha256"
	flags["node"] = fmt.Sprintf("%v/%v", n.Number, n.TotalNodes)

	var johnPath string
	if len(os.Getenv("JOHN_PATH")) == 0 {
		johnPath = "john"
	} else {
		johnPath = os.Getenv("JOHN_PATH")
	}

	cmd := john.New(
		johnPath,
		"dummy",
		flags,
		logger,
	)

	go func() {
		msgs, _ := n.Start(cmd)
		for {
			select {
			case msg := <-msgs:
				sugar.Debug("main got %v", msg)
			}
		}
	}()

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
	sugar.Infof("stopping node %v", n.Number)

}
