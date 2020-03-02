package crd

import (
	"strings"
	"testing"
)

func TestResourceNames(t *testing.T) {
	tbl := []struct {
		project string
		robot   string
		out     string
		err     string
	}{
		{
			project: "voo-faa",
			robot:   "foo$bar/baz",
			out:     "voo-faa-foo-bar-baz-16283",
		},
		{
			project: "team-foo-bar-baz-aaaasdasdasdasdasd",
			robot:   "foo$barbaz[]/=FFFasdasdasdsd",
			out:     "",
			err:     "resource name too long",
		},
		// the following three should not collide
		{
			project: "team-foo",
			robot:   "foo$barbaz[]=FFF",
			out:     "team-foo-foo-barbaz-FFF-26883",
		},
		{
			project: "team-foo",
			robot:   "foo$barbaz[]/=FFF",
			out:     "team-foo-foo-barbaz-FFF-28141",
		},
		{
			project: "team",
			robot:   "foo-foo$barbaz[]/=FFF",
			out:     "team-foo-foo-barbaz-FFF-27825",
		},
	}

	for i, item := range tbl {
		out, err := buildResourceName(item.project, item.robot)
		if err != nil && !strings.Contains(err.Error(), item.err) {
			t.Errorf("[%d] expected err %s, found %s", i, item.err, err)
		}
		if out != item.out {
			t.Errorf("[%d] expected %s, found %s", i, item.out, out)
		}

	}
}
