module github.com/isbm/uyuni-ncd

go 1.13

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/uuid v1.1.1
	github.com/isbm/go-nanoconf v0.0.0-20200213162501-c88ba6d6d64c
	github.com/lib/pq v1.3.0
	github.com/nats-io/nats-server/v2 v2.1.4 // indirect
	github.com/nats-io/nats.go v1.9.1
	github.com/urfave/cli/v2 v2.1.1
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678
)

replace github.com/isbm/go-nanoconf => /home/bo/work/golang/go-nanoconf
