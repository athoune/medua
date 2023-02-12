all: medusa-proxy medusa-get medusa-wal

medusa-proxy: bin
	go build -o bin/medusa-proxy cli/medusa-proxy/proxy.go

medusa-get: bin
	go build -o bin/medusa-get cli/medusa-get/mget.go

medusa-wal: bin
	go build -o bin/medusa-wal cli/medusa-wal/mwal.go

bin:
	mkdir -p bin

test:
	go test -v -cover ./multiclient

clean:
	rm -rf bin
