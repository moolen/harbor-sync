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

package controllers

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeSecret(namespace, name string, baseURL string, credentials crdv1.RobotAccountCredential) v1.Secret {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", credentials.Name, credentials.Token)))
	configJSON := fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","auth":"%s"}}}`, baseURL, credentials.Name, credentials.Token, auth)
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Type: v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: []byte(configJSON),
		},
	}
}

func upsertSecret(cl client.Client, log logr.Logger, secret v1.Secret) {
	err := cl.Create(context.Background(), &secret)
	if apierrs.IsAlreadyExists(err) {
		err = cl.Update(context.TODO(), &secret)
		if err != nil {
			log.Error(err, "could not update secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
			return
		}
		log.Info("updated secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
		return
	}
	if err != nil {
		log.Error(err, "could not create secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
		return
	}
	log.Info("created secret", "proposed_namespace", secret.ObjectMeta.Namespace, "proposed_secret", secret.ObjectMeta.Name)
	return
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}
