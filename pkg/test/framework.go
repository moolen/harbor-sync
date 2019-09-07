/*
Copyright 2019 The Authors.

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

package test

import (
	"context"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureHarborSyncConfig(cl client.Client, name string) crdv1.HarborSync {
	cfg := &crdv1.HarborSync{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: name},
		Spec: crdv1.HarborSyncSpec{
			Type:               crdv1.RegexMatching,
			ProjectName:        name,
			RobotAccountSuffix: "sync-bot",
			Mapping:            []crdv1.ProjectMapping{},
		},
		Status: crdv1.HarborSyncStatus{
			RobotCredentials: map[string]crdv1.RobotAccountCredential{},
		},
	}
	err := cl.Create(context.Background(), cfg)
	if !apierrs.IsAlreadyExists(err) {
		Expect(err).ToNot(HaveOccurred())
	}
	return *cfg
}

func EnsureHarborSyncConfigWithParams(cl client.Client, name, projectName string, mapping *crdv1.ProjectMapping, whc []crdv1.WebhookConfig) crdv1.HarborSync {
	var mappings []crdv1.ProjectMapping

	if mapping != nil {
		mappings = append(mappings, *mapping)
	}

	cfg := &crdv1.HarborSync{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: crdv1.HarborSyncSpec{
			Type:               crdv1.RegexMatching,
			ProjectName:        projectName,
			RobotAccountSuffix: "sync-bot",
			Mapping:            mappings,
			Webhook:            whc,
		},
		Status: crdv1.HarborSyncStatus{
			RobotCredentials: map[string]crdv1.RobotAccountCredential{},
		},
	}
	err := cl.Create(context.Background(), cfg)
	if !apierrs.IsAlreadyExists(err) {
		Expect(err).ToNot(HaveOccurred())
	}
	return *cfg
}

func DeleteHarborSyncConfig(cl client.Client, name string) {
	err := cl.Delete(context.Background(), &crdv1.HarborSync{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}, client.GracePeriodSeconds(0))
	Expect(err).ToNot(HaveOccurred())
}

func EnsureNamespace(cl client.Client, namespace string) {
	err := cl.Create(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})
	if apierrs.IsAlreadyExists(err) {
		return
	}
	Expect(err).ToNot(HaveOccurred())
}

func DeleteNamespace(cl client.Client, namespace string) {
	err := cl.Delete(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, client.GracePeriodSeconds(0))
	if apierrs.IsConflict(err) {
		return
	}
	Expect(err).ToNot(HaveOccurred())
}

func DeleteSecret(cl client.Client, ns, name string) {
	err := cl.Delete(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}, client.GracePeriodSeconds(0))

	Expect(err).ToNot(HaveOccurred())
}
