package disk

import (
	"io/ioutil"
	"testing"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
)

func TestStore(t *testing.T) {
	tmp, err := ioutil.TempDir("", "diskstore")
	if err != nil {
		t.Error(err)
	}
	t.Logf("store tmp dir: %s", tmp)
	c, err := New(tmp)
	if err != nil {
		t.Error(err)
	}
	// doesn't exist
	_, err = c.Get("foo", "bar")
	if err == nil {
		t.Errorf("expected failure")
	}
	// set project foo cred
	err = c.Set("foo", crdv1.RobotAccountCredential{
		Name:  "robot$foo",
		Token: "1234",
	})
	if err != nil {
		t.Error(err)
	}
	// set project bar cred
	err = c.Set("bar", crdv1.RobotAccountCredential{
		Name:  "robot$bar",
		Token: "2345",
	})
	if err != nil {
		t.Error(err)
	}

	// check foo
	exists := c.Has("foo", "robot$foo")
	if exists == false {
		t.Errorf("expected key to exist")
	}
	cred, err := c.Get("foo", "robot$foo")
	if err != nil {
		t.Error(err)
	}
	if cred.Name != "robot$foo" {
		t.Errorf("unexpected robot name")
	}
	if cred.Token != "1234" {
		t.Errorf("unexpected robot token")
	}

	// check bar
	exists = c.Has("bar", "robot$bar")
	if exists == false {
		t.Errorf("expected key to exist")
	}
	cred, err = c.Get("bar", "robot$bar")
	if err != nil {
		t.Error(err)
	}
	if cred.Name != "robot$bar" {
		t.Errorf("unexpected robot name")
	}
	if cred.Token != "2345" {
		t.Errorf("unexpected robot token")
	}

	k := c.Keys()
	if k[0][0] != "bar" {
		t.Errorf("unexpected key")
	}
	if k[0][1] != "robot$bar" {
		t.Errorf("unexpected key")
	}
	if k[1][0] != "foo" {
		t.Errorf("unexpected key")
	}
	if k[1][1] != "robot$foo" {
		t.Errorf("unexpected key")
	}

	//
	// delete all entries
	//
	err = c.Reset()
	if err != nil {
		t.Error(err)
	}
	// check foo
	exists = c.Has("foo", "robot$foo")
	if exists == true {
		t.Errorf("expected key to not exist")
	}
}
