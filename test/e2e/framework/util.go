/*
Copyright 2014 The Kubernetes Authors.
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
package framework

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	// Poll how often to poll for conditions
	Poll = 2 * time.Second

	// DefaultTimeout time to wait for operations to complete
	DefaultTimeout = 90 * time.Second
)

// RunID unique identifier of the e2e run
var RunID = uuid.NewUUID()

// CreateNamespace creates a new namespace in the cluster
func CreateNamespace(name string, c kubernetes.Interface) (string, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	// Be robust about making the namespace creation call.
	var got *corev1.Namespace
	var err error

	err = wait.PollImmediate(Poll, DefaultTimeout, func() (bool, error) {
		got, err = c.CoreV1().Namespaces().Create(ns)
		if err != nil {
			log.Errorf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return got.Name, nil
}

// CreateKubeNamespace creates a new namespace in the cluster
func CreateKubeNamespace(baseName string, c kubernetes.Interface) (string, error) {
	ts := time.Now().UnixNano()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("e2e-tests-%v-%v-", baseName, ts),
		},
	}
	// Be robust about making the namespace creation call.
	var got *corev1.Namespace
	var err error

	err = wait.PollImmediate(Poll, DefaultTimeout, func() (bool, error) {
		got, err = c.CoreV1().Namespaces().Create(ns)
		if err != nil {
			log.Errorf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return got.Name, nil
}

// DeleteKubeNamespace deletes a namespace and all the objects inside
func DeleteKubeNamespace(c kubernetes.Interface, namespace string) error {
	grace := int64(0)
	pb := metav1.DeletePropagationBackground
	return c.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &pb,
	})
}

// WaitForKubeNamespaceNotExist waits until a namespaces is not present in the cluster
func WaitForKubeNamespaceNotExist(c kubernetes.Interface, namespace string) error {
	return wait.PollImmediate(Poll, DefaultTimeout, namespaceNotExist(c, namespace))
}

func namespaceNotExist(c kubernetes.Interface, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := c.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}
}
