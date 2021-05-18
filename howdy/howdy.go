package howdy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

type Server struct {
	res       Results
	hold      sync.Mutex
	port      int
	sessionID string
	log       *zap.SugaredLogger
	etcd      *clientv3.Client
}

type Results struct {
	Nodes map[int]Node `json:"node"`
}

type Node struct {
	Status  string   `json:"status"`
	Results []string `json:"results"`
}

func New(port int, logger *zap.Logger, etcd *clientv3.Client) Server {
	return Server{
		port: port,
		log:  logger.Sugar(),
		etcd: etcd,
	}
}

func (s *Server) ValidSession() bool {
	re, err := s.etcd.KV.Get(context.TODO(), "session/id")
	if err != nil {
		s.log.Error(err)
		return false
	}
	if len(re.Kvs) == 0 {
		s.log.Warn("no active session found")
		return false
	}
	if string(re.Kvs[0].Value) != s.sessionID {
		s.log.Infof("found new session %v", string(re.Kvs[0].Value))
		s.sessionID = string(re.Kvs[0].Value)
	}
	return true
}

func (s *Server) GetCurrent() Results {
	re, _ := s.etcd.KV.Get(context.TODO(), "session/"+s.sessionID, clientv3.WithPrefix())
	if len(re.Kvs) == 0 {
		s.res = Results{}
		return Results{}
	}

	r := map[int]Node{}
	for _, kv := range re.Kvs {
		key := string(kv.Key)
		value := kv.Value
		p := strings.Split(key, "/")
		node := Node{}
		nn, err := strconv.Atoi(p[len(p)-1])
		if err == nil {
			node.Status = "alive"
		}
		if p[len(p)-1] == "results" {
			nn, _ = strconv.Atoi(p[len(p)-2])
			d := []string{}
			err := json.Unmarshal(value, &d)
			if err != nil {
				s.log.Error(err)
			}
			node.Status = "alive"
			node.Results = d
		}
		r[nn] = node
	}
	return Results{Nodes: r}
}

func (s *Server) Serve() {
	go func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				var j []byte
				if s.ValidSession() {
					j, _ = json.Marshal(s.GetCurrent())
				} else {
					j, _ = json.Marshal(map[string]string{"error": "no active session"})
					s.log.Warn("found no valid session")
				}
				w.Write(j)
				s.log.Debugf("served %v", string(j))
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "nothing here dude")
				s.log.Info("bad request")
			}
		}
		http.HandleFunc("/", handler)
		s.log.Infof("serving on port %v", s.port)
		http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil)
	}()
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
	s.log.Info("stopping howdy")
}
