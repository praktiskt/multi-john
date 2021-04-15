package howdy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

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
	Nodes map[string]Node `json:"nodes"`
}

type Node struct {
	Results []string `json:"results"`
}

func New(port int, sessionID string, logger *zap.Logger, etcd *clientv3.Client) Server {
	return Server{
		port:      port,
		sessionID: sessionID,
		log:       logger.Sugar(),
		etcd:      etcd,
	}
}

func (s *Server) CheckSession() {
	//TODO: Check if session is still alive. If not, detatch.
}

func (s *Server) GetCurrent() {
	re, _ := s.etcd.KV.Get(context.TODO(), "session/"+s.sessionID, clientv3.WithPrefix())
	if len(re.Kvs) == 0 {
		s.res = Results{}
		return
	}

	r := map[string]Node{}
	for _, kv := range re.Kvs {
		key := string(kv.Key)
		value := kv.Value
		p := strings.Split(key, "/")
		if p[len(p)-1] == "results" {
			d := []string{}
			nn := p[len(p)-2]
			err := json.Unmarshal(value, &d)
			if err != nil {
				s.log.Error(err)
			}
			r[nn] = Node{Results: d}
		}
	}
	s.res = Results{Nodes: r}
}

func (s *Server) Serve() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			s.GetCurrent()
			j, _ := json.Marshal(s.res)
			w.Write(j)
			s.log.Infof("served %v", string(j))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "nothing here dude")
			s.log.Info("bad request")
		}
	}
	http.HandleFunc("/", handler)
	s.log.Infof("serving on port %v", s.port)
	http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil)
}
