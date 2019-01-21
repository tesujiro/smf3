test:
	go test -v ./data/db

restart:
	kill -9 `ps -ef | grep tile38 | grep -v grep |awk '{print $$2}'` && tile38-server &

watch:
	while : ;do netstat -an|grep 9851 | wc -l; sleep 2 ;done
