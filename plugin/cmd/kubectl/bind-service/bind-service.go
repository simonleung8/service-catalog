/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"os"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/plugin/cmd/kubectl/utils"

	"k8s.io/client-go/pkg/api/v1"
)

const usage = `Usage:
  kubectl plugin bind-service INSTANCE_NAME BINDING_NAME [--namespace]`

func main() {
	if len(os.Args) != 3 {
		utils.Exit1(usage)
	}

	scClient, config := utils.NewClient()

	binding := v1alpha1.ServiceInstanceCredential{}
	binding.Kind = "binding"
	binding.Name = os.Args[2]
	binding.Namespace = utils.Namespace()
	binding.Spec.ServiceInstanceRef = v1.LocalObjectReference{
		Name: os.Args[1],
	}
	binding.Spec.SecretName = os.Args[2]

	utils.CheckNamespaceExists(binding.Namespace, config)
	utils.Ok()

	fmt.Printf("Creating binding %s to %s in Namespace %s...\n",
		utils.Entity(binding.Name),
		utils.Entity(binding.Spec.ServiceInstanceRef.Name),
		utils.Entity(binding.Namespace))
	resp, err := scClient.ServiceInstanceCredentials(binding.Namespace).Create(&binding)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error binding service instance (%s)", err))
	}
	utils.Ok()

	table := utils.NewTable("BINDING NAME", "NAMESPACE", "INSTANCE NAME", "SECRET NAME")
	table.AddRow(resp.Name, resp.Namespace, resp.Spec.ServiceInstanceRef.Name, resp.Spec.SecretName)
	err = table.Print()
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
	}
}
