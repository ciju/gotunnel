all: 
	go build -ldflags "-X main.version `cat VERSION`" -o gotunnel client.go
