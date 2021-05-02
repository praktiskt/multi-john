package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/magnusfurugard/multi-john/howdy"
	"github.com/magnusfurugard/multi-john/worker"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

var mode string

func init() {
	flag.StringVar(&mode, "mode", "worker", "mode to start in, must be worker or howdy")
}

func main() {
	// Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	flag.Parse()
	sugar.Info(mode)

	// Find etcd
	endpoint := []string{}
	if s, ok := os.LookupEnv("ETCD_ADVERTISE_CLIENT_URLS"); ok {
		endpoint = append(endpoint, strings.Split(s, ",")...)
	} else {
		endpoint = append(endpoint, "localhost:2379")
		//sugar.Panicf("found no advertised client urls for etcd")
	}

	sugar.Info("connect to etcd...")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		sugar.Panic(err)
	}
	defer cli.Close()

	// Configure howdy
	if mode == "howdy" {

		s := howdy.New(8080, logger, cli)
		s.Serve()
	} else {
		// Start worker
		node := worker.New(logger, cli)
		// Wait for termination signal
		termChan := make(chan os.Signal)
		signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
		<-termChan
		sugar.Infof("stopping node %v", node.Number)
	}

}
