# WIP

## Instructions
make build
make docker-build
make cluster

## wait for everything to be ready:
kubectl get nodes
kubectl -n kube-system get pods

## deploy admission controller
make deploy
