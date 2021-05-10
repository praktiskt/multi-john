package worker

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/magnusfurugard/multi-john/worker/john"
	"github.com/magnusfurugard/multi-john/worker/node"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

func New(logger *zap.Logger, cli *clientv3.Client, johnFile string, johnFlags string) node.Node {
	sugar := logger.Sugar()

	// Cofigure node
	var totalNodes int
	if n, ok := os.LookupEnv("TOTAL_NODES"); ok {
		totalNodes, _ = strconv.Atoi(n)
	} else {
		sugar.Warn("TOTAL_NODES environment missing, defaulting to TOTAL_NODES=2")
		totalNodes = 2
	}

	n, err := node.New(totalNodes, cli, logger)
	if err != nil {
		sugar.Errorf("Unable to start node: %v", err)
		cli.Close()
		os.Exit(1)
	}

	// Configure john
	flags := map[string]string{}
	fl := strings.Split(johnFlags, ",")
	if len(fl[0]) > 0 {
		for _, flag := range fl {
			f := strings.SplitN(flag, "=", 2)
			flags[f[0]] = f[1]
			sugar.Info(flags)
		}
	}
	flags["--node"] = fmt.Sprintf("%v/%v", n.Number, n.TotalNodes)

	var johnPath string
	if j, ok := os.LookupEnv("JOHN_PATH"); ok {
		johnPath = j
	} else {
		johnPath = "john"
	}

	cmd := john.New(
		johnPath,
		johnFile,
		flags,
		logger,
	)

	// Start john on node
	go func() {
		msgs := n.Start(cmd)
		for {
			select {
			case msg := <-msgs:
				if len(msg.Events) == 0 {
					continue
				}
				sugar.Infof("msg %v", string(msg.Events[0].Kv.Value))
			}
		}
	}()

	return *n
}
