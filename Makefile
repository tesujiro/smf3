test:
	go vet ./...
	go test -v ./data/db ./web

.PHONY: cover
cover:
	go vet ./...
	go test -v ./data/db ./web -coverpkg ./... -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

restart:
	kill -9 `ps -ef | grep tile38 | grep -v grep |awk '{print $$2}'` && tile38-server &

.PHONY: rambler
rambler: ./web/rambler/*.go
	go build -o rambler ./web/rambler

.PHONY: web
web: ./web/*.go
	#go build -o smfweb ./web
	go build -o ./web/smfweb ./web

size:
	find . -name \*.go -or -name \*.js | grep -v _test | xargs wc -l

watch:
	while : ;do netstat -an|grep 9851 | wc -l; sleep 2 ;done

lsof:
	while : ;do lsof -p `pgrep -n tile38-server`| wc -l; sleep 2 ;done

socket:
	while : ;do printf  "socket(port:9851) %d \tfile descriptor %d\n" `netstat -an|grep 9851 | wc -l` $$(lsof -p `pgrep -n tile38-server`| wc -l); sleep 2 ;done
