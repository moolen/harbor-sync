package controllers

import (
	"context"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureHarborSyncConfig(cl client.Client, name string) {
	err := cl.Create(context.Background(), &crdv1.HarborSyncConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: name},
		Spec: crdv1.HarborSyncConfigSpec{
			ProjectSelector: []crdv1.ProjectSelector{},
		},
		Status: crdv1.HarborSyncConfigStatus{
			RobotCredentials: map[string]crdv1.RobotAccountCredentials{},
		},
	})
	if !apierrs.IsAlreadyExists(err) {
		Expect(err).ToNot(HaveOccurred())
	}
}

func ensureHarborSyncConfigWithParams(cl client.Client, name, projectName string, mapping crdv1.ProjectMapping) {
	err := cl.Create(context.Background(), &crdv1.HarborSyncConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: crdv1.HarborSyncConfigSpec{
			ProjectSelector: []crdv1.ProjectSelector{
				{
					Type:               crdv1.RegexMatching,
					ProjectName:        projectName,
					RobotAccountSuffix: "sync-bot",
					Mapping: []crdv1.ProjectMapping{
						mapping,
					},
				},
			},
		},
		Status: crdv1.HarborSyncConfigStatus{
			RobotCredentials: map[string]crdv1.RobotAccountCredentials{},
		},
	})
	if !apierrs.IsAlreadyExists(err) {
		Expect(err).ToNot(HaveOccurred())
	}
}

func deleteHarborSyncConfig(cl client.Client, name string) {
	err := cl.Delete(context.Background(), &crdv1.HarborSyncConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	})
	Expect(err).ToNot(HaveOccurred())
}

func ensureNamespace(cl client.Client, namespace string) {
	cl.Create(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})
}

func deleteNamespace(cl client.Client, namespace string) {
	cl.Delete(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})
}
