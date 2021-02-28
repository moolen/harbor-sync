package crd

import (
	"context"
	"fmt"
	"regexp"
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

var reg = regexp.MustCompile("[^a-zA-Z0-9-]+")

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
	rname, err := BuildResourceName(project, name)
	if err != nil {
		return nil, err
	}
	var r crdv1.HarborRobotAccount
	err = s.kubeClient.Get(ctx, types.NamespacedName{Name: rname}, &r)
	if err != nil {
		log.Errorf("error getting robot account %s: %s", rname, err)
		return nil, err
	}
	return &r.Spec.Credential, nil
}

func (s *Store) Set(project string, cred crdv1.RobotAccountCredential) error {
	ctx := context.Background()
	rname, err := BuildResourceName(project, cred.Name)
	if err != nil {
		return err
	}
	r := crdv1.HarborRobotAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: rname,
			Annotations: map[string]string{
				"project": project,
				"robot":   cred.Name,
			},
		},
		Spec: crdv1.HarborRobotAccountSpec{
			Credential: cred,
		},
		Status: crdv1.HarborRobotAccountStatus{
			LastSync: time.Now().Unix(),
		},
	}
	err = s.kubeClient.Create(ctx, &r)
	if apierrs.IsAlreadyExists(err) {
		err = s.kubeClient.Get(ctx, types.NamespacedName{Name: rname}, &r)
		if err != nil {
			return fmt.Errorf("could not update robot account %s: %s", rname, err)
		}
		r.Spec.Credential = cred
		r.Status = crdv1.HarborRobotAccountStatus{
			LastSync: time.Now().Unix(),
		}
		err = s.kubeClient.Update(context.Background(), &r)
		if err != nil {
			return fmt.Errorf("could not update robot account %s: %s", rname, err)
		}
		return nil
	}
	return err
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

func BuildResourceName(project, robot string) (string, error) {
	// 2.2.0 introduces "global" robot accounts
	// when using the old API they get created
	// with a different name: robot${project-name}+{provided-name}
	// on the GET side we map them back to robot${provided-name}
	robot = strings.TrimPrefix(robot, fmt.Sprintf("robot$%s+", project))
	in := fmt.Sprintf("%s-%s-%d", project, robot, hash(project, robot))
	out := reg.ReplaceAllString(in, "-")
	if len(out) > 63 {
		return "", fmt.Errorf("resource name too long (%s / %s): %s", project, robot, out)
	}
	return out, nil
}

func hash(project, robot string) int {
	h := 0
	r := []rune(fmt.Sprintf("%s|%s", project, robot))
	for i, n := range r {
		h += i * int(n)
	}
	return h
}
