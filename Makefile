.PHONY: test
test:
	@echo "\n🛠️  Running unit tests..."
	go test ./...

.PHONY: build
build:
	@echo "\n🔧  Building Go binaries..."
	GOOS=darwin GOARCH=amd64 go build -o bin/admission-webhook-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -o bin/admission-webhook-linux-amd64 .

.PHONY: docker-build
docker-build:
	@echo "\n📦 Building legit-security Docker image..."
	cp /tmp/cosign.pub key.pub
	DOCKER_BUILDKIT=1 docker build -t legit-security:latest .

# From this point `kind` is required
.PHONY: cluster
cluster:
	@echo "\n🔧 Creating Kubernetes cluster..."
	kind create cluster --config dev/manifests/kind/kind.cluster.yaml

.PHONY: delete-cluster
delete-cluster:
	@echo "\n♻️  Deleting Kubernetes cluster..."
	kind delete cluster

.PHONY: push
push: docker-build
	@echo "\n📦 Pushing admission-webhook image into Kind's Docker daemon..."
	kind load docker-image legit-security:latest

.PHONY: deploy-config
deploy-config:
	@echo "\n⚙️  Applying cluster config..."
	kubectl apply -f dev/manifests/cluster-config/

.PHONY: delete-config
delete-config:
	@echo "\n♻️  Deleting Kubernetes cluster config..."
	kubectl delete -f dev/manifests/cluster-config/

.PHONY: deploy
deploy: push delete deploy-config
	@echo "\n🚀 Deploying legit-security..."
	kubectl apply -f dev/manifests/webhook/

.PHONY: delete
delete:
	@echo "\n♻️  Deleting legit-security deployment if existing..."
	kubectl delete -f dev/manifests/webhook/ || true

.PHONY: pod
pod:
	@echo "\n🚀 Deploying test pod..."
	kubectl apply -f dev/manifests/pods/hello-world.pod.yaml

.PHONY: delete-pod
delete-pod:
	@echo "\n♻️ Deleting test pod..."
	kubectl delete -f dev/manifests/pods/hello-world.pod.yaml

.PHONY: no-provenance
no-provenance:
	@echo "\n🚀 Deploying \"no-provenance\" pod..."
	kubectl apply -f dev/manifests/pods/bad-name.pod.yaml

.PHONY: delete-no-provenance
delete-no-provenance:
	@echo "\n🚀 Deleting \"no-provenance\" pod..."
	kubectl delete -f dev/manifests/pods/bad-name.pod.yaml

.PHONY: bad-digest
bad-digest:
	@echo "\n🚀 Deploying \"bad-digest\" pod..."
	kubectl apply -f dev/manifests/pods/bad-digest.pod.yaml

.PHONY: delete-bad-digest
delete-bad-digest:
	@echo "\n🚀 Deleting \"bad-digest\" pod..."
	kubectl delete -f dev/manifests/pods/bad-digest.pod.yaml

.PHONY: taint
taint:
	@echo "\n🎨 Taining Kubernetes node.."
	kubectl taint nodes kind-control-plane "acme.com/lifespan-remaining"=4:NoSchedule

.PHONY: logs
logs:
	@echo "\n🔍 Streaming legit-security logs..."
	kubectl logs -l app=legit-security -f

.PHONY: delete-all
delete-all: delete delete-config delete-pod delete-bad-pod
