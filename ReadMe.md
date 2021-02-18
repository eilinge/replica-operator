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
5. kubectl -n repcontroller-system get pod
```
repcontroller-controller-manager-7d9d4fdb84-mwxh7   2/2     Running   0          32s
```

## Issue

1. kube-rbac-proxy pull failed

    replace config/default/proxy.yaml image to `jimmysong/kubebuilder-kube-rbac-proxy:v0.5.0`

2. controller manager image pull failed

    add yaml: `imagePullPolicy: IfNotPresent`

3. remind: **donnot use `go mod tidy` will due to run failed**
