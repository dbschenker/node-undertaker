all: clean test build

build: build-node-undertaker

build-node-undertaker:
	go build -o bin/node-undertaker gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker

test: mock
	go test ./...

clean:
	rm -r bin/ || true

docker:
	docker buildx build -t node-undertaker:local .

lint:
	golangci-lint run ./... -v

mock:
	go generate ./...

clean_mocks:
	find . -name '*_mocks.go' -delete

vuln:
	govulncheck ./...

vet:
	go vet ./...

kind:
	kind create cluster --config example/kind/config.yaml

local:
	bin/node-undertaker --namespace node-undertaker --log-level=debug --cloud-provider=kwok --cloud-termination-delay=60 --drain-delay=60

kwok:
	kwok create cluster
	kubectl config use-context kwok-kwok