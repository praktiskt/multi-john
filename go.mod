module github.com/magnusfurugard/multi-john

go 1.16

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	go.uber.org/atomic => github.com/uber-go/atomic v1.5.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0

)

require (
	github.com/coreos/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/google/uuid v1.2.0
	github.com/prometheus/client_golang v1.10.0 // indirect
	go.etcd.io/etcd v3.3.25+incompatible
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/tools v0.1.0 // indirect
)
