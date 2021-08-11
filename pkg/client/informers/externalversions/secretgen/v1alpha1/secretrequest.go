// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	secretgenv1alpha1 "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen/v1alpha1"
	versioned "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/client/clientset/versioned"
	internalinterfaces "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/client/listers/secretgen/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// SecretRequestInformer provides access to a shared informer and lister for
// SecretRequests.
type SecretRequestInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.SecretRequestLister
}

type secretRequestInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewSecretRequestInformer constructs a new informer for SecretRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSecretRequestInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredSecretRequestInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredSecretRequestInformer constructs a new informer for SecretRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredSecretRequestInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SecretgenV1alpha1().SecretRequests(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SecretgenV1alpha1().SecretRequests(namespace).Watch(options)
			},
		},
		&secretgenv1alpha1.SecretRequest{},
		resyncPeriod,
		indexers,
	)
}

func (f *secretRequestInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredSecretRequestInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *secretRequestInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&secretgenv1alpha1.SecretRequest{}, f.defaultInformer)
}

func (f *secretRequestInformer) Lister() v1alpha1.SecretRequestLister {
	return v1alpha1.NewSecretRequestLister(f.Informer().GetIndexer())
}
