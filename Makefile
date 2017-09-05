test:
	go test github.com/elwin0214/gomemcached/demo -v
	go test github.com/elwin0214/gomemcached/logger -v
	go test github.com/elwin0214/gomemcached/sock -v
	go test github.com/elwin0214/gomemcached/memcached -v
	go test github.com/elwin0214/gomemcached/util -v

bench:
	go test -bench . -run=^Benchmark github.com/elwin0214/gomemcached/logger -v
	go test -bench . -run=^Benchmark github.com/elwin0214/gomemcached/sock -v
	go test -bench . -run=^Benchmark github.com/elwin0214/gomemcached/memcached -v

build:
	go build -o main/mem_client main/mem_client.go
	go build -o main/mem_server main/mem_server.go
	go build -o main/raw_bench main/raw_bench.go

clean:
	rm main/mem_client
	rm main/mem_server
	rm main/raw_bench
