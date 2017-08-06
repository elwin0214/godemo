
test:
	go test src/demo/*_test.go -v
	go test src/logger/*_test.go -v
	go test src/sock/*_test.go -v
	go test src/memcached/*_test.go -v 
	go test src/util/*_test.go -v 

bench:
	go test -bench . -run=^Benchmark src/logger/*_test.go -v
	go test -bench . -run=^Benchmark src/sock/*_test.go -v
	go test -bench . -run=^Benchmark src/memcached/*_test.go -v 

map_bench:
	go test src/logger/demo/bench_map_test.go
	
install:
	go install src/godemo/mem_client.go
	go install src/godemo/mem_server.go
	go install src/godemo/raw_bench.go

clean:
	rm -rf ./bin
	rm -rf ./pkg