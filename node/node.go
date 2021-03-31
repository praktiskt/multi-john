package node

import (
	"context"
	"fmt"
	"log"

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
	var s string

	re, err := n.etcd.KV.Get(context.TODO(), sID)
	if err != nil {
		return err
	}

	if len(re.Kvs) == 0 {
		n.SessionID = uuid.NewString()
		tx := n.etcd.Txn(context.TODO())
		_, err := tx.If().Then(
			clientv3.OpPut(sID, s),
		).Commit()
		if err != nil {
			return err
		}
	} else {
		n.SessionID = string(re.Kvs[0].Value)
	}

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

func (n *Node) Start(johnCmd john.Cmd) {
	johnCmd.Run()
}
