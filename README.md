[![Build Status](https://travis-ci.org/inwinstack/ipam-operator.svg?branch=master)](https://travis-ci.org/inwinstack/ipam-operator) [![Docker Build Status](https://img.shields.io/docker/build/inwinstack/ipam-operator.svg)](https://hub.docker.com/r/inwinstack/ipam-operator/) [![codecov](https://codecov.io/gh/inwinstack/ipam-operator/branch/master/graph/badge.svg)](https://codecov.io/gh/inwinstack/ipam-operator) ![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)
# IPAM Operator
An operator to assign IP for Kubernetes Namespace. This operator will provide two custom resource(Pool and IP).

![](images/architecture.png)

## Building from Source
Clone repo into your go path under `$GOPATH/src`:
```sh
$ git clone https://github.com/inwinstack/ipam-operator.git $GOPATH/src/github.com/inwinstack/ipam-operator
$ cd $GOPATH/src/github.com/inwinstack/ipam-operator
$ make dep
$ make
```

## Debug out of the cluster
Run the following command to debug:
```sh
$ go run cmd/main.go \
    --kubeconfig $HOME/.kube/config \
    --default-ignore-namespaces=kube-system,default,kube-public \
    --default-address=192.168.100.0/24 \
    --logtostderr -v=2
```

## Deploy in the cluster
Run the following command to deploy operator:
```sh
$ kubectl apply -f deploy/
$ kubectl -n kube-system get po
```
