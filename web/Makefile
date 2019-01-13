MODULE=sfweb
SSL_KEY=ssl/server.key
SSL_CSR=ssl/server.csr
SSL_CRT=ssl/server.crt

all: build 

build: ./*.go
	go build -o $(MODULE)
	
.PHONY: deploy
deploy:
	for p in `ps -ef | grep $(MODULE) | awk '{print $$2}'`;do\
		kill -9 $$p;\
	done
	./$(MODULE)

create_secret:
ifndef APIKEY
        $(error APIKEY not set)
endif
	echo $$APIKEY > ./apikey.txt
	kubectl create secret generic google-apikey --from-file=apikey=./apikey.txt
	rm ./apikey.txt
	
test:
	go test . -v

logs:
	#kubectl logs `kubectl get pod | grep $(MODULE) | awk '{print $$1}'` --follow

ssl_certificate:
	openssl genrsa 2048 > $(SSL_KEY)
	openssl req -new -key $(SSL_KEY) > $(SSL_CSR)
	openssl x509 -days 3650 -req -signkey $(SSL_KEY) < $(SSL_CSR) > $(SSL_CRT)
