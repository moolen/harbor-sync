package store

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	"github.com/peterbourgon/diskv/v3"
)

// DiskStore is a cache on-disk
type DiskStore struct {
	path string
	c    *diskv.Diskv
}

func transformFunc(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last],
	}
}

func inverseTransform(pathKey *diskv.PathKey) (key string) {
	return path.Join(strings.Join(pathKey.Path, "/"), pathKey.FileName)
}

// New returns a new DiskStore
func New(path string) (*DiskStore, error) {
	c := diskv.New(diskv.Options{
		BasePath:          path,
		AdvancedTransform: transformFunc,
		InverseTransform:  inverseTransform,
		CacheSizeMax:      1024 * 1024,
	})

	return &DiskStore{
		path: path,
		c:    c,
	}, nil
}

// NewTemp returns a new DiskStore
func NewTemp() (*DiskStore, error) {
	tmp, err := ioutil.TempDir("", "diskstore")
	if err != nil {
		return nil, err
	}
	return New(tmp)
}

// Has returns true if the item exists, false if not
func (d *DiskStore) Has(project, name string) bool {
	return d.c.Has(path.Join(project, name))
}

// Get returns a item
func (d *DiskStore) Get(project, name string) (*crdv1.RobotAccountCredential, error) {
	var cred crdv1.RobotAccountCredential
	rd, err := d.c.ReadStream(path.Join(project, name), true)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(rd)
	err = json.Unmarshal(data, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// Set writes data to the disk store
func (d *DiskStore) Set(project string, cred crdv1.RobotAccountCredential) error {
	err := os.MkdirAll(path.Join(d.path, project), os.ModePerm)
	if err != nil {
		return err
	}
	data, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	return d.c.Write(path.Join(project, cred.Name), data)
}

// Keys returns all available keys
func (d *DiskStore) Keys() [][]string {
	var keys [][]string
	files, err := ioutil.ReadDir(d.path)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.IsDir() && f.Name() != "." {
			kc := d.c.KeysPrefix(f.Name(), nil)
			for key := range kc {
				segments := strings.Split(key, "/")
				keys = append(keys, []string{segments[0], segments[1]})
			}
		}
	}
	return keys
}

// Reset deletes all entries from the cache
func (d *DiskStore) Reset() error {
	return d.c.EraseAll()
}
