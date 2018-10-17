[![Build Status](https://travis-ci.org/inwinstack/ipam-operator.svg?branch=master)](https://travis-ci.org/inwinstack/ipam-operator) ![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)
# IPAM Operator
An operator to auto assign IP for Kubernetes Namespace and Service.

# Building from Source
Clone repo into your go path under `$GOPATH/src`:
```sh
$ git clone https://github.com/inwinstack/ipam-operator.git $GOPATH/src/github.com/inwinstack/ipam-operator
$ cd $GOPATH/src/github.com/inwinstack/ipam-operator
$ make dep
$ make
```

# Debug out of the cluster
Run the following command to debug:
```sh
$ go run cmd/main.go \
    --kubeconfig $HOME/.kube/config \
    --logtostderr 
```

# Deploy in the cluster
Run the following command to deploy operator:
```sh
$ kubectl apply -f deploy/
$ kubectl -n kube-system get po
```