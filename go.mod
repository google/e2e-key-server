module github.com/google/keytransparency

go 1.12

require (
	cloud.google.com/go/spanner v1.2.0
	github.com/VividCortex/mysqlerr v0.0.0-20170204212430-6c6b55f8796f
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.4.2
	github.com/google/certificate-transparency-go v1.1.0 // indirect
	github.com/google/go-cmp v0.4.1
	github.com/google/tink/go v1.4.0-rc2
	github.com/google/trillian v1.3.9-0.20200515134453-00fd1983ab83
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/kr/pretty v0.1.0
	github.com/kylelemons/godebug v1.1.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/prometheus/client_golang v1.5.1
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	go.uber.org/atomic v1.5.1 // indirect
	golang.org/x/crypto v0.0.0-20191119213627-4f8c1d86b1ba
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/api v0.24.0
	google.golang.org/genproto v0.0.0-20200331122359-1ee6d9798940
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20200513171258-e048e166ab9c
