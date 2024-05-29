all: clean test build

build: build-node-undertaker

build-node-undertaker:
	go build -o bin/node-undertaker github.com/dbschenker/node-undertaker/cmd/node-undertaker

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
	bin/node-undertaker --namespace kube-node-lease --log-level=debug --cloud-provider=kwok --cloud-termination-delay=180 --cloud-prepare-termination-delay=200 --drain-delay=190 --node-initial-threshold 45

kwok:
	kwokctl create cluster --node-lease-duration-seconds 0
	kubectl config use-context kwok-kwok
	kwokctl get kubeconfig > ~/.kube/kwok.kubeconfig
