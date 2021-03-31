package node

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/magnusfurugard/multi-john/john"
	"go.etcd.io/etcd/clientv3"
)

var (
	s   = "session"
	sID = s + "/id"

	timeoutSeconds = int64(2)
)

type Node struct {
	etcd       *clientv3.Client
	Number     int
	TotalNodes int
	SessionID  string
}

func path(args ...string) string {
	var s string
	for _, arg := range args {
		s += arg + "/"
	}
	return s[:len(s)-1]
}

func New(maxNodes int, etcd *clientv3.Client) (*Node, error) {
	n := Node{}
	n.etcd = etcd
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
	//var s string
	newID := uuid.NewString()

	lease := clientv3.NewLease(n.etcd)
	timeout, err := lease.Grant(context.TODO(), timeoutSeconds*10)
	if err != nil {
		return err
	}

	p := path(sID)
	re, err := n.etcd.KV.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.CreateRevision(p), ">", 0)).
		Then(clientv3.OpGet(p)).
		Else(clientv3.OpPut(p, newID, clientv3.WithLease(timeout.ID))).
		Commit()

	if re.Responses[0].GetResponsePut() != nil {
		n.SessionID = newID
		return nil
	}

	res, err := n.etcd.KV.Get(context.TODO(), sID)
	if err != nil {
		return err
	}
	n.SessionID = string(res.Kvs[0].Value)

	return nil
}

func (n *Node) GetNodeNumber() error {
	lease := clientv3.NewLease(n.etcd)
	timeout, err := lease.Grant(context.TODO(), timeoutSeconds)
	if err != nil {
		return err
	}

	for i := 1; i <= n.TotalNodes; i++ {
		p := path(s, n.SessionID, "node", fmt.Sprint(i))
		re, err := n.etcd.KV.Txn(context.TODO()).
			If(clientv3.Compare(clientv3.Value(p), "=", "taken")).
			Then(clientv3.OpGet(p)).
			Else(clientv3.OpPut(p, "taken", clientv3.WithLease(timeout.ID))).
			Commit()
		if err != nil {
			return err
		}
		if re.Responses[0].GetResponsePut() != nil {
			n.Number = i
			log.Printf("i am now node %v", i)
			return nil
		}
	}
	return fmt.Errorf("no more nodes are available")
}

func (n *Node) GetTotalNodes() int {
	return 2
}

func (n *Node) Heartbeat() error {
	lease := clientv3.NewLease(n.etcd)
	timeout, err := lease.Grant(context.TODO(), timeoutSeconds)
	if err != nil {
		return err
	}
	p := path(s, n.SessionID, "node", fmt.Sprint(n.Number))
	_, err = n.etcd.KV.Put(context.TODO(), p, "taken", clientv3.WithLease(timeout.ID))
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) KeepAlive() chan bool {
	ticker := time.NewTicker(time.Duration(timeoutSeconds) * time.Second)
	var kill chan bool
	go func() {
		for {
			select {
			case <-kill:
				break
			case <-ticker.C:
				n.Heartbeat()
			}
		}
	}()
	return kill
}

func (n *Node) Start(johnCmd john.Cmd) []string {
	kill := n.KeepAlive()
	johnCmd.Run()
	kill <- true
	return johnCmd.Results()
}
