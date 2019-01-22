test:
	go vet ./...
	go test -v ./data/db ./web

restart:
	kill -9 `ps -ef | grep tile38 | grep -v grep |awk '{print $$2}'` && tile38-server &

size:
	find . -name \*.go -or -name \*.js | grep -v _test | xargs wc -l

watch:
	while : ;do netstat -an|grep 9851 | wc -l; sleep 2 ;done
