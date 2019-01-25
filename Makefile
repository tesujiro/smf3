test:
	go vet ./...
	go test -v ./data/db ./web

restart:
	kill -9 `ps -ef | grep tile38 | grep -v grep |awk '{print $$2}'` && tile38-server &

rambler: ./web/rambler/*.go
	go build -o rambler ./web/rambler

.PHONY: web
web: ./web/*.go
	go build -o smfweb ./web

size:
	find . -name \*.go -or -name \*.js | grep -v _test | xargs wc -l

watch:
	while : ;do netstat -an|grep 9851 | wc -l; sleep 2 ;done
