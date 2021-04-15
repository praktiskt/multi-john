package node

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/magnusfurugard/multi-john/worker/john"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

var (
	timeoutSeconds = int64(10)
)

type Node struct {
	etcd       *clientv3.Client
	Log        *zap.SugaredLogger
	Number     int
	SessionID  string
	TotalNodes int
	Paths      struct {
		SessionID string
		Node      string
		Results   string
	}
}

type Msg struct {
	TS      time.Time
	Payload string
}

func path(args ...string) string {
	var s string
	for _, arg := range args {
		s += arg + "/"
	}
	return s[:len(s)-1]
}

func New(maxNodes int, etcd *clientv3.Client, logger *zap.Logger) (*Node, error) {
	n := Node{}
	n.etcd = etcd
	n.Log = logger.Sugar()
	n.Paths.SessionID = path("session", "id")
	if err := n.GetSession(); err != nil {
		return nil, err
	}
	n.TotalNodes = maxNodes
	if err := n.GetNodeNumber(); err != nil {
		return nil, err
	}

	return &n, nil
}

func (n *Node) GetSession() error {
	newID := uuid.NewString()
	re, err := n.etcd.KV.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.CreateRevision(n.Paths.SessionID), ">", 0)).
		Then(clientv3.OpGet(n.Paths.SessionID)).
		Else(clientv3.OpPut(n.Paths.SessionID, newID)).
		Commit()

	if err != nil {
		return err
	}

	if re.Responses[0].GetResponsePut() != nil {
		n.SessionID = newID
		n.Log.Debugf("created new session:", n.SessionID)
	} else {
		res, err := n.etcd.KV.Get(context.TODO(), n.Paths.SessionID)
		if err != nil {
			return err
		}
		n.SessionID = string(res.Kvs[0].Value)
		n.Log.Debugf("found existing session:", n.SessionID)
	}
	n.Log.Info("connected to session %v", n.SessionID)
	go n.keepAlive(n.Paths.SessionID, n.SessionID)
	return nil
}

func (n *Node) GetNodeNumber() error {
	lease := clientv3.NewLease(n.etcd)
	timeout, err := lease.Grant(context.TODO(), timeoutSeconds)
	if err != nil {
		return err
	}

	for i := 1; i <= n.TotalNodes; i++ {
		p := path("session", n.SessionID, "node", fmt.Sprint(i))
		re, err := n.etcd.KV.Txn(context.TODO()).
			If(clientv3.Compare(clientv3.Value(p), "=", "taken")).
			Then(clientv3.OpGet(p)).
			Else(clientv3.OpPut(p, "taken", clientv3.WithLease(timeout.ID))).
			Commit()
		if err != nil {
			return err
		}
		if re.Responses[0].GetResponsePut() != nil {
			n.Log.Infof("i am node %v", i)
			n.Number = i

			n.Paths.Node = p
			n.Paths.Results = path(p, "results")
			return nil
		}
	}
	return fmt.Errorf("no more nodes are available")
}

func (n *Node) heartbeat(key string, value string) error {
	lease := clientv3.NewLease(n.etcd)
	timeout, err := lease.Grant(context.TODO(), timeoutSeconds)
	if err != nil {
		return err
	}
	_, err = n.etcd.KV.Put(context.TODO(), key, value, clientv3.WithLease(timeout.ID))
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) keepAlive(key string, value string) {
	jitter := func() int64 {
		rand.Seed(time.Now().UnixNano())
		return rand.Int63n(timeoutSeconds / 2)
	}
	for {
		n.Log.Debugf("sending heartbeat for %v=%v", key, value)
		if err := n.heartbeat(key, value); err != nil {
			n.Log.Error(err)
		}
		time.Sleep(time.Duration(timeoutSeconds-jitter()) * time.Second)
	}
}

func (n *Node) writeResults(results chan []string) {
	for {
		select {
		case msgs := <-results:
			found, err := json.Marshal(msgs)
			if err != nil {
				n.Log.Error(err)
			}
			n.Log.Debugf("etcd put %v=%v", n.Paths.Results, string(found))
			_, err = n.etcd.KV.Put(context.TODO(), n.Paths.Results, string(found))
			if err != nil {
				n.Log.Error(err)
			}
		}
	}
}

func (n *Node) Start(johnCmd john.Cmd) clientv3.WatchChan {
	go n.keepAlive(n.Paths.Node, "taken")
	go johnCmd.Run()
	go n.writeResults(johnCmd.Results)
	return n.etcd.Watch(context.TODO(), n.Paths.Results)
}
