/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/starizard/kube-envoy-controller/pkg/api/example.com/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// EnvoyLister helps list Envoys.
type EnvoyLister interface {
	// List lists all Envoys in the indexer.
	List(selector labels.Selector) (ret []*v1.Envoy, err error)
	// Envoys returns an object that can list and get Envoys.
	Envoys(namespace string) EnvoyNamespaceLister
	EnvoyListerExpansion
}

// envoyLister implements the EnvoyLister interface.
type envoyLister struct {
	indexer cache.Indexer
}

// NewEnvoyLister returns a new EnvoyLister.
func NewEnvoyLister(indexer cache.Indexer) EnvoyLister {
	return &envoyLister{indexer: indexer}
}

// List lists all Envoys in the indexer.
func (s *envoyLister) List(selector labels.Selector) (ret []*v1.Envoy, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Envoy))
	})
	return ret, err
}

// Envoys returns an object that can list and get Envoys.
func (s *envoyLister) Envoys(namespace string) EnvoyNamespaceLister {
	return envoyNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// EnvoyNamespaceLister helps list and get Envoys.
type EnvoyNamespaceLister interface {
	// List lists all Envoys in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.Envoy, err error)
	// Get retrieves the Envoy from the indexer for a given namespace and name.
	Get(name string) (*v1.Envoy, error)
	EnvoyNamespaceListerExpansion
}

// envoyNamespaceLister implements the EnvoyNamespaceLister
// interface.
type envoyNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Envoys in the indexer for a given namespace.
func (s envoyNamespaceLister) List(selector labels.Selector) (ret []*v1.Envoy, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Envoy))
	})
	return ret, err
}

// Get retrieves the Envoy from the indexer for a given namespace and name.
func (s envoyNamespaceLister) Get(name string) (*v1.Envoy, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("envoy"), name)
	}
	return obj.(*v1.Envoy), nil
}
