package broker

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
)

const (
	session         = "current-session"
	sessionID       = session + "/id"
	sessionLastSeen = sessionID + "/last_seen"
)

var (
	heartbeatTimeoutSeconds = 5
	heartbeat               = time.NewTicker(time.Duration(1) * time.Second)
)

type Broker struct {
	Etcd    *clientv3.Client
	Session Session
}

type Session struct {
	ID   string
	Kill chan bool
}

func Connect(etcd *clientv3.Client, nodes int) (*Broker, error) {
	session, err := getOrCreateSession(etcd)
	if err != nil {
		return nil, err
	}

	/*
		TODO: See if there is an existing session
		TODO: Set active broker session id
		TODO: Set session max nodes
		TODO: Set next node to hand out
	*/
	return &Broker{Etcd: etcd, Session: *session}, nil
}

func (b *Broker) GetNumberOfNodes() int {
	// TODO: fetch number of nodes
	return 2
}

func (b *Broker) GetAvailableNodeSlot() int {
	//TODO: Fetch the next available node slot, and set as occupied
	return 1
}

func stillHere(etcd *clientv3.Client) error {
	re, err := etcd.Get(context.TODO(), sessionLastSeen)
	if err != nil {
		return err
	}

	var lastSeen int64
	now := time.Now().Unix()
	for _, kv := range re.Kvs {
		lastSeen, err = strconv.ParseInt(string(kv.Value), 10, 64)
		if err != nil {
			return err
		}
	}

	if lastSeen < (now - int64(heartbeatTimeoutSeconds)) {
		_, err := etcd.Put(context.TODO(), sessionLastSeen, fmt.Sprint(now))
		if err != nil {
			return err
		}
	}
	return nil
}

func getOrCreateSession(etcd *clientv3.Client) (*Session, error) {
	var s Session

	re, err := etcd.Get(context.TODO(), sessionID)
	if err != nil {
		return nil, err
	}

	if len(re.Kvs) == 0 {
		s.ID = uuid.NewString()
		tx := etcd.Txn(context.TODO())
		_, err := tx.If().Then(
			clientv3.OpPut(sessionID, s.ID),
		).Commit()
		if err != nil {
			return nil, err
		}
	} else {
		for _, kv := range re.Kvs {
			st := string(kv.Value)
			s.ID = st
		}
	}

	// Keep session alive
	go func() {
		for {
			select {
			case <-s.Kill:
				log.Print("killing")
				break
			case <-heartbeat.C:
				re, _ := etcd.Get(context.TODO(), sessionLastSeen)
				log.Print(re.Kvs[0])
				stillHere(etcd)
			}
		}
	}()

	return &s, nil
}
