This branch inspired by https://github.com/inwinstack/ipam .


The IPAM provides `Pool` and `IP` custom resource to take care of assigning and unassigning individual addresses from pools because Kubernetes cannot create IP addresses out of thin air, so we need to give it CRDs that it can use.

![](images/architecture.png)

## Building from Source
Clone repo into your go path under `$GOPATH/src`:
```sh
$ git clone https://github.com/xenolog/ipam $GOPATH/src/github.com/inwinstack/ipam
$ cd $GOPATH/src/github.com/inwinstack/ipam
$ make
```

## Debug out of the cluster
Run the following command to debug:
```sh
$ go run cmd/main.go \
    --kubeconfig $HOME/.kube/config \
    --logtostderr \
    -v=2
```

## Deploy in the cluster
Run the following command to deploy the controller:
```sh
$ kubectl apply -f deploy/
$ kubectl -n kube-system get po -l ipam
```
