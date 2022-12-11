all: medusa-proxy medusa-get

medusa-proxy: bin
	go build -o bin/medusa-proxy cli/medusa-proxy/proxy.go

medusa-get: bin
	go build -o bin/medusa-get cli/medusa-get/mget.go

bin:
	mkdir -p bin

clean:
	rm -rf bin
