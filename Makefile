FILE=VERSION
USERNAME=aniruddhabasak
PROJECTNAME=kontroller
IMAGE=${USERNAME}/${PROJECTNAME}
TAG=$(shell cat VERSION)

.PHONY: increment-minor
increment-minor:
	python3 version.py --increment-minor

.PHONY: docker-no-increment
docker-no-increment:
	docker build -t ${IMAGE}:${TAG} .
	docker push ${IMAGE}:${TAG}
	yq e -i '(.spec.template.spec.containers[] | select(.image) | .image) |= "${IMAGE}:${TAG}"' manifests/deploy.yaml

.PHONY: docker
docker: increment-patch
	docker build -t ${IMAGE}:${TAG} .
	docker push ${IMAGE}:${TAG}
	yq e -i '(.spec.template.spec.containers[] | select(.image) | .image) |= "${IMAGE}:${TAG}"' manifests/deploy.yaml

.PHONY: increment-patch
increment-patch:
	python3 version.py
	@TAG=$(shell cat ${FILE})

.PHONY: docker-dev
docker-dev:
	docker-compose up -d

.PHONY: ssl
ssl:
	rm -f manifests/certs/tls.crt manifests/certs/tls.key
	openssl req -x509 -nodes -days 365 -newkey rsa:4096 -keyout manifests/certs/tls.key -out manifests/certs/tls.crt -config manifests/certs/tls.cnf -extensions 'v3_req'
	kubectl create secret generic certs --from-file manifests/certs/tls.crt --from-file manifests/certs/tls.key --dry-run=client -o yaml > manifests/secret.yaml

.PHONY: e2e-ssl
e2e-ssl:
	rm -f test/e2e/testdata/certs/tls.crt test/e2e/testdata/certs/tls.key
	openssl req -x509 -nodes -days 365 -newkey rsa:4096 -keyout test/e2e/testdata/certs/tls.key -out test/e2e/testdata/certs/tls.crt -config test/e2e/testdata/certs/tls-e2e.cnf -extensions 'v3_req'
	kubectl create secret generic certs --from-file test/e2e/testdata/certs/tls.crt --from-file test/e2e/testdata/certs/tls.key --dry-run=client -o yaml > test/e2e/testdata/secret.yaml

.PHONY: manifest
manifest:
	kubectl apply -f manifests/
	kubectl apply -f manifests/admission/

.PHONY: drop-manifest
drop-manifest:
	kubectl delete -f manifests/
	kubectl delete -f manifests/admission/

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

.PHONY: e2e-test
e2e-test:
	@go mod tidy
	go test -v -tags e2e ./test/e2e

.PHONY: lint
lint:
	gofmt -s -w .
	golangci-lint run
