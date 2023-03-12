FILE=VERSION
USERNAME=aniruddhabasak
PROJECTNAME=kontroller
IMAGE=${USERNAME}/${PROJECTNAME}
TAG=$(shell cat VERSION)

.PHONY: docker
docker: docker-build
	docker push ${IMAGE}:${TAG}
	yq e -i '(.spec.template.spec.containers[] | select(.image) | .image) |= "${IMAGE}:${TAG}"' manifests/deploy.yaml

.PHONY: docker-build
docker-build: pre-docker
	docker build -t ${IMAGE}:${TAG} .

.PHONY: pre-docker
pre-docker:
	python3 version.py
	@TAG=$(shell cat ${FILE})

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
