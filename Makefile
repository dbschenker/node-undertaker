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

kind_load:
	kind load docker-image node-undertaker:local

kind_helm:
	helm upgrade --install -n node-undertaker node-undertaker charts/node-undertaker --create-namespace -f example/kind/values.yaml

local:
	bin/node-undertaker --namespace kube-node-lease --log-level=debug --cloud-provider=kwok --cloud-termination-delay=45 --drain-delay=90 --node-initial-threshold 45

kwok:
	kwok create cluster
	kubectl config use-context kwok-kwok

local_chart:
	helm package charts/node-undertaker -d ../gilds-node-undertaker/chart/node-undertaker/charts/
