MODULE=rambler

all: build

build: ./*.go
	go build -o $(MODULE) 

.PHONY: bench
bench: 
	go test -bench . -benchmem

.PHONY: profile
profile:
	go test -v . -cpuprofile cpu.out
	go tool pprof --svg $(MODULE).test cpu.out > cpu.svg

