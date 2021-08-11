// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PasswordLister helps list Passwords.
type PasswordLister interface {
	// List lists all Passwords in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Password, err error)
	// Passwords returns an object that can list and get Passwords.
	Passwords(namespace string) PasswordNamespaceLister
	PasswordListerExpansion
}

// passwordLister implements the PasswordLister interface.
type passwordLister struct {
	indexer cache.Indexer
}

// NewPasswordLister returns a new PasswordLister.
func NewPasswordLister(indexer cache.Indexer) PasswordLister {
	return &passwordLister{indexer: indexer}
}

// List lists all Passwords in the indexer.
func (s *passwordLister) List(selector labels.Selector) (ret []*v1alpha1.Password, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Password))
	})
	return ret, err
}

// Passwords returns an object that can list and get Passwords.
func (s *passwordLister) Passwords(namespace string) PasswordNamespaceLister {
	return passwordNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PasswordNamespaceLister helps list and get Passwords.
type PasswordNamespaceLister interface {
	// List lists all Passwords in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Password, err error)
	// Get retrieves the Password from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Password, error)
	PasswordNamespaceListerExpansion
}

// passwordNamespaceLister implements the PasswordNamespaceLister
// interface.
type passwordNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Passwords in the indexer for a given namespace.
func (s passwordNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Password, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Password))
	})
	return ret, err
}

// Get retrieves the Password from the indexer for a given namespace and name.
func (s passwordNamespaceLister) Get(name string) (*v1alpha1.Password, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("password"), name)
	}
	return obj.(*v1alpha1.Password), nil
}
