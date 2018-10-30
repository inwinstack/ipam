package main

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/inwinstack/ipam-operator/pkg/operator"
	"github.com/inwinstack/ipam-operator/pkg/version"
	flag "github.com/spf13/pflag"
)

var (
	kubeconfig       string
	address          string
	namespaces       []string
	autoAssign       bool
	ignoreAnnotation bool
	ver              bool
)

func parserFlags() {
	flag.StringVarP(&kubeconfig, "kubeconfig", "", "", "Absolute path to the kubeconfig file.")
	flag.StringVarP(&address, "default-address", "", "", "Set default IP pool address.")
	flag.StringSliceVarP(&namespaces, "default-ignore-namespaces", "", nil, "Set default IP pool ignore namespaces.")
	flag.BoolVarP(&autoAssign, "default-auto-assign", "", true, "Set default IP pool ignore namespace annotation.")
	flag.BoolVarP(&ignoreAnnotation, "default-ignore-annotation", "", false, "Set default IP pool ignore namespace annotation.")
	flag.BoolVarP(&ver, "version", "", false, "Display the version")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
}

func main() {
	defer glog.Flush()
	parserFlags()

	if ver {
		fmt.Fprintf(os.Stdout, "%s\n", version.GetVersion())
		os.Exit(0)
	}

	glog.Infof("Starting IPAM operator...")

	f := &operator.Flag{
		Kubeconfig:                kubeconfig,
		IgnoreNamespaces:          namespaces,
		Address:                   address,
		AutoAssignToNamespace:     autoAssign,
		IgnoreNamespaceAnnotation: ignoreAnnotation,
	}

	op := operator.NewMainOperator(f)
	if err := op.Initialize(); err != nil {
		glog.Fatalf("Error initing operator instance: %v.\n", err)
	}

	if err := op.Run(); err != nil {
		glog.Fatalf("Error serving operator instance: %s.\n", err)
	}
}
