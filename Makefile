.PHONY: docker
docker:
	docker build -t aniruddhabasak/kontroller:latest .

.PHONY: docker-push
docker-push:
	docker push aniruddhabasak/kontroller:latest


.PHONY: docker-dev
docker-dev:
	docker-compose up

.PHONY: ssl
ssl:
	rm -f manifests/certs/secret.yaml manifests/certs/tls.crt manifests/certs/tls.key
	openssl req -x509 -nodes -days 365 -newkey rsa:4096 -keyout manifests/certs/tls.key -out manifests/certs/tls.crt -config manifests/certs/tls.cnf -extensions 'v3_req'
	kubectl create secret generic certs --from-file manifests/certs/tls.crt --from-file manifests/certs/tls.key --dry-run=client -o yaml > manifests/secret.yaml

.PHONY: build
build:
	go build -o ./bin/kontroller ./main.go
	gofmt -d .

.PHONY: run
run:
	go run main.go -kubeconfig=$(HOME)/.kube/config

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	gofmt -s -w .
	golangci-lint run