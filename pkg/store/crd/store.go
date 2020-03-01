package crd

import (
	"context"
	"fmt"
	"strings"
	"time"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	log "github.com/sirupsen/logrus"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Store struct {
	kubeClient client.Client
}

// New returns a new Store
func New(kubeClient client.Client) (*Store, error) {
	return &Store{
		kubeClient: kubeClient,
	}, nil
}

func (s *Store) Has(project, name string) bool {
	_, err := s.Get(project, name)
	if err != nil {
		return false
	}
	return true
}

func (s *Store) Get(project, name string) (*crdv1.RobotAccountCredential, error) {
	ctx := context.Background()
	rname := buildResourceName(project, name)
	var r crdv1.HarborRobotAccount
	err := s.kubeClient.Get(ctx, types.NamespacedName{Name: rname}, &r)
	if err != nil {
		log.Errorf("error getting robot account %s: %s", rname, err)
		return nil, err
	}
	return &r.Spec.Credential, nil
}

func (s *Store) Set(project string, cred crdv1.RobotAccountCredential) error {
	ctx := context.Background()
	rname := buildResourceName(project, cred.Name)
	r := crdv1.HarborRobotAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: rname,
		},
		Spec: crdv1.HarborRobotAccountSpec{
			Credential: cred,
		},
		Status: crdv1.HarborRobotAccountStatus{
			LastSync: time.Now().Unix(),
		},
	}
	err := s.kubeClient.Create(ctx, &r)
	if apierrs.IsAlreadyExists(err) {
		err = s.kubeClient.Get(ctx, types.NamespacedName{Name: rname}, &r)
		if err != nil {
			return fmt.Errorf("could not update robot account %s: %s", rname, err)
		}
		r.Spec.Credential = cred
		r.Status = crdv1.HarborRobotAccountStatus{
			LastSync: time.Now().Unix(),
		}
		err = s.kubeClient.Update(context.TODO(), &r)
		if err != nil {
			return fmt.Errorf("could not update robot account %s: %s", rname, err)
		}
		return nil
	}
	return fmt.Errorf("could not create robot account %s: %s", rname, err)
}

func (s *Store) Reset() error {
	ctx := context.Background()
	r := crdv1.HarborRobotAccount{}
	err := s.kubeClient.DeleteAllOf(ctx, &r)
	if err != nil {
		return err
	}
	return nil
}

func buildResourceName(project, robot string) string {
	in := fmt.Sprintf("%s-%s", project, robot)
	return strings.ReplaceAll(in, "$", "-")
}
