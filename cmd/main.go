package main

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/inwinstack/ipam-operator/pkg/operator"
	flag "github.com/spf13/pflag"
)

var (
	kubeconfig string
)

func parserFlags() {
	flag.StringVarP(&kubeconfig, "kubeconfig", "", "", "Absolute path to the kubeconfig file.")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
}

func main() {
	defer glog.Flush()
	parserFlags()

	glog.Infof("Starting IPAM operator...")

	f := &operator.Flag{Kubeconfig: kubeconfig}
	op := operator.NewMainOperator(f)
	if err := op.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initing operator instance: %s.\n", err)
		os.Exit(1)
	}

	if err := op.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error serving operator instance: %s.\n", err)
		os.Exit(1)
	}
}
