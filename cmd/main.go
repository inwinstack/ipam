/*
Copyright © 2018 inwinSTACK.inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	goflag "flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/inwinstack/ipam/pkg/operator"
	"github.com/inwinstack/ipam/pkg/version"
	flag "github.com/spf13/pflag"
)

var (
	kubeconfig string
	ver        bool
)

func parserFlags() {
	flag.StringVarP(&kubeconfig, "kubeconfig", "", "", "Absolute path to the kubeconfig file.")
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

	glog.Infof("Starting IPAM controller...")

	op := operator.NewMainOperator(kubeconfig)
	if err := op.Initialize(); err != nil {
		glog.Fatalf("Error initing operator instance: %+v.\n", err)
	}

	if err := op.Run(); err != nil {
		glog.Fatalf("Error serving operator instance: %+v.\n", err)
	}
}
