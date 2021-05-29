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
	"go.uber.org/zap/zapcore"
)

var mode string
var logLevel string
var johnFile string
var johnFlags string

func init() {
	flag.StringVar(&mode, "mode", "worker", "mode to start in, must be worker or howdy")
	flag.StringVar(&logLevel, "logLevel", "info", "log level, info or debug")
	flag.StringVar(&johnFile, "johnFile", "dummy", "the file with hashes to process")
	flag.StringVar(&johnFlags, "johnFlags", "", "a comma-separated list of flags to pass, e.g. --format=raw-sha256,--fork=2")
}

func GetLogLevel() zapcore.Level {
	if strings.ToLower(logLevel) == "debug" {
		return zapcore.DebugLevel
	}
	return zapcore.InfoLevel
}

func main() {
	flag.Parse()
	// Logger
	logger, _ := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(GetLogLevel()),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}.Build()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

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

	if mode == "howdy" {
		// Configure howdy
		s := howdy.New(8080, logger, cli)
		go s.Serve()
	} else {
		// Start worker
		if err := worker.New(logger, cli, johnFile, johnFlags); err != nil {
			sugar.Panic(err)
		}
	}
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
}
