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

package injector

import (
	"encoding/json"
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
)

type k8sBindingInjector struct {
	client kubernetes.Interface
}

// CreateK8sBindingInjector creates an instance of a BindingInjector which
// manages the injection of binding information within the Kubernetes
// environment.
func CreateK8sBindingInjector() (BindingInjector, error) {
	// This assumes that we are running withing a kubernetes cluster. If this
	// needs to be able to run outside the cluster, it will need to be modified
	// to take a non-default config.
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &k8sBindingInjector{
		client: client,
	}, nil
}

func (b *k8sBindingInjector) Inject(binding *servicecatalog.Binding, cred *brokerapi.Credential) error {
	secret := &v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      binding.Name,
			Namespace: binding.Spec.InstanceRef.Namespace,
		},
		Data: make(map[string][]byte),
	}

	// For each item in the cred just serialize its value into JSON and
	// save it in the secret
	for k, v := range *cred {
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("Unable to marshal credential value %q: %s",
				k, err)
		}
		secret.Data[k] = data
	}
	secretsCl := b.client.Core().Secrets(binding.Spec.InstanceRef.Namespace)
	_, err := secretsCl.Create(secret)
	return err
}

func (b *k8sBindingInjector) Uninject(binding *servicecatalog.Binding) error {
	secretsCl := b.client.Core().Secrets(binding.Namespace)
	gracePeriodSec := int64(0)
	return secretsCl.Delete(binding.Name, &api.DeleteOptions{
		TypeMeta:           unversioned.TypeMeta{Kind: "DeleteOptions"},
		GracePeriodSeconds: &gracePeriodSec,
	})
}
