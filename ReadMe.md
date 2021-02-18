# replica-operator

This project is to control the number of replicas of a specific deployment

## require

install kustomize controller-gen

## Deployment process

make && make install && make run
kubectl apply -f config/samples/batch_v1_controller.yaml
kubectl get controller.batch.controller.kubebuilder.io
