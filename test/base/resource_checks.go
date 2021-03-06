/*
Copyright 2019 The Knative Authors

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

// crdpolling contains functions which poll Knative Serving CRDs until they
// get into the state desired by the caller or time out.

package base

import (
	"context"
	"fmt"
	"time"

	"github.com/knative/pkg/apis/duck"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"go.opencensus.io/trace"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
)

// MetaResource includes necessary meta data to retrieve the duck-type KResource.
type MetaResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// Meta returns a MetaResource built from the given name, namespace and kind.
func Meta(name, namespace, kind string) *MetaResource {
	return &MetaResource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: EventingAPIVersion,
		},
	}
}

const (
	// The interval and timeout used for polling in checking resource states.
	interval = 1 * time.Second
	timeout  = 4 * time.Minute
)

// WaitForResourceReady polls the status of the MetaResource from client
// every interval until isResourceReady returns `true` indicating
// it is done, returns an error or timeout. desc will be used to
// name the metric that is emitted to track how long it took for
// the resource to get into the state checked by isResourceReady.
func WaitForResourceReady(dynamicClient dynamic.Interface, obj *MetaResource) error {
	metricName := fmt.Sprintf("WaitForResourceReady/%s", obj.Name)
	_, span := trace.StartSpan(context.Background(), metricName)
	defer span.End()

	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		return isResourceReady(dynamicClient, obj)
	})
}

// isResourceReady leverage duck-type to check if the given MetaResource is in ready state
func isResourceReady(dynamicClient dynamic.Interface, obj *MetaResource) (bool, error) {
	// get the resource's name, namespace and gvr
	name := obj.Name
	namespace := obj.Namespace
	gvk := obj.GroupVersionKind()
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	// use the helper functions to convert the resource to a KResource duck
	tif := &duck.TypedInformerFactory{Client: dynamicClient, Type: &duckv1alpha1.KResource{}}
	_, lister, err := tif.Get(gvr)
	if err != nil {
		// Return error to stop the polling.
		return false, err
	}
	untyped, err := lister.ByNamespace(namespace).Get(name)
	if k8serrors.IsNotFound(err) {
		// Return false as we are not done yet.
		// We swallow the error to keep on polling.
		// It should only happen if we wait for the auto-created resources, like default Broker.
		return false, nil
	} else if err != nil {
		// Return error to stop the polling.
		return false, err
	}
	kr := untyped.(*duckv1alpha1.KResource)
	return kr.Status.GetCondition(duckv1alpha1.ConditionReady).IsTrue(), nil
}
