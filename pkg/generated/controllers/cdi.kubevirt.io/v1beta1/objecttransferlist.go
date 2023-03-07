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

// Code generated by main. DO NOT EDIT.

package v1beta1

import (
	"context"
	"time"

	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/pkg/generic"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	v1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

type ObjectTransferListHandler func(string, *v1beta1.ObjectTransferList) (*v1beta1.ObjectTransferList, error)

type ObjectTransferListController interface {
	generic.ControllerMeta
	ObjectTransferListClient

	OnChange(ctx context.Context, name string, sync ObjectTransferListHandler)
	OnRemove(ctx context.Context, name string, sync ObjectTransferListHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() ObjectTransferListCache
}

type ObjectTransferListClient interface {
	Create(*v1beta1.ObjectTransferList) (*v1beta1.ObjectTransferList, error)
	Update(*v1beta1.ObjectTransferList) (*v1beta1.ObjectTransferList, error)

	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1beta1.ObjectTransferList, error)
	List(namespace string, opts metav1.ListOptions) (*v1beta1.ObjectTransferListList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ObjectTransferList, err error)
}

type ObjectTransferListCache interface {
	Get(namespace, name string) (*v1beta1.ObjectTransferList, error)
	List(namespace string, selector labels.Selector) ([]*v1beta1.ObjectTransferList, error)

	AddIndexer(indexName string, indexer ObjectTransferListIndexer)
	GetByIndex(indexName, key string) ([]*v1beta1.ObjectTransferList, error)
}

type ObjectTransferListIndexer func(obj *v1beta1.ObjectTransferList) ([]string, error)

type objectTransferListController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewObjectTransferListController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) ObjectTransferListController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &objectTransferListController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromObjectTransferListHandlerToHandler(sync ObjectTransferListHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1beta1.ObjectTransferList
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1beta1.ObjectTransferList))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *objectTransferListController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1beta1.ObjectTransferList))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateObjectTransferListDeepCopyOnChange(client ObjectTransferListClient, obj *v1beta1.ObjectTransferList, handler func(obj *v1beta1.ObjectTransferList) (*v1beta1.ObjectTransferList, error)) (*v1beta1.ObjectTransferList, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *objectTransferListController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *objectTransferListController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), handler))
}

func (c *objectTransferListController) OnChange(ctx context.Context, name string, sync ObjectTransferListHandler) {
	c.AddGenericHandler(ctx, name, FromObjectTransferListHandlerToHandler(sync))
}

func (c *objectTransferListController) OnRemove(ctx context.Context, name string, sync ObjectTransferListHandler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), FromObjectTransferListHandlerToHandler(sync)))
}

func (c *objectTransferListController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *objectTransferListController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *objectTransferListController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *objectTransferListController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *objectTransferListController) Cache() ObjectTransferListCache {
	return &objectTransferListCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *objectTransferListController) Create(obj *v1beta1.ObjectTransferList) (*v1beta1.ObjectTransferList, error) {
	result := &v1beta1.ObjectTransferList{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *objectTransferListController) Update(obj *v1beta1.ObjectTransferList) (*v1beta1.ObjectTransferList, error) {
	result := &v1beta1.ObjectTransferList{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *objectTransferListController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *objectTransferListController) Get(namespace, name string, options metav1.GetOptions) (*v1beta1.ObjectTransferList, error) {
	result := &v1beta1.ObjectTransferList{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *objectTransferListController) List(namespace string, opts metav1.ListOptions) (*v1beta1.ObjectTransferListList, error) {
	result := &v1beta1.ObjectTransferListList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *objectTransferListController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *objectTransferListController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v1beta1.ObjectTransferList, error) {
	result := &v1beta1.ObjectTransferList{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type objectTransferListCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *objectTransferListCache) Get(namespace, name string) (*v1beta1.ObjectTransferList, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1beta1.ObjectTransferList), nil
}

func (c *objectTransferListCache) List(namespace string, selector labels.Selector) (ret []*v1beta1.ObjectTransferList, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ObjectTransferList))
	})

	return ret, err
}

func (c *objectTransferListCache) AddIndexer(indexName string, indexer ObjectTransferListIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1beta1.ObjectTransferList))
		},
	}))
}

func (c *objectTransferListCache) GetByIndex(indexName, key string) (result []*v1beta1.ObjectTransferList, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1beta1.ObjectTransferList, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1beta1.ObjectTransferList))
	}
	return result, nil
}
