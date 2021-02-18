# replica-operator

This project is to control the number of replicas of a specific deployment

## require

install kustomize controller-gen

## Deployment process

### local run

1. make && make install && make run
2. kubectl apply -f config/samples/batch_v1_controller.yaml
3. kubectl get controller.batch.controller.kubebuilder.io

### deploy k8s

1. make
2. make docker-build
3. make docker-push
4. make deploy
