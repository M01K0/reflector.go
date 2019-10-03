package store

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/lbryio/lbry.go/extras/errors"
)

// FileBlobStore is a local disk store.
type FileBlobStore struct {
	// the location of blobs on disk
	blobDir string
	// store files in subdirectories based on the first N chars in the filename. 0 = don't create subdirectories.
	prefixLength int

	initialized bool
}

// NewFileBlobStore returns an initialized file disk store pointer.
func NewFileBlobStore(dir string, prefixLength int) *FileBlobStore {
	return &FileBlobStore{blobDir: dir, prefixLength: prefixLength}
}

func (f *FileBlobStore) dir(hash string) string {
	if f.prefixLength <= 0 || len(hash) < f.prefixLength {
		return f.blobDir
	}
	return path.Join(f.blobDir, hash[:f.prefixLength])
}

func (f *FileBlobStore) path(hash string) string {
	return path.Join(f.dir(hash), hash)
}

func (f *FileBlobStore) ensureDirExists(dir string) error {
	return errors.Err(os.MkdirAll(dir, 0755))
}

func (f *FileBlobStore) initOnce() error {
	if f.initialized {
		return nil
	}

	err := f.ensureDirExists(f.blobDir)
	if err != nil {
		return err
	}

	f.initialized = true
	return nil
}

// Has returns T/F or Error if it the blob stored already. It will error with any IO disk error.
func (f *FileBlobStore) Has(hash string) (bool, error) {
	err := f.initOnce()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(f.path(hash))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Get returns the byte slice of the blob stored or will error if the blob doesn't exist.
func (f *FileBlobStore) Get(hash string) ([]byte, error) {
	err := f.initOnce()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(f.path(hash))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Err(ErrBlobNotFound)
		}
		return nil, err
	}

	return ioutil.ReadAll(file)
}

// Put stores the blob on disk
func (f *FileBlobStore) Put(hash string, blob []byte) error {
	err := f.initOnce()
	if err != nil {
		return err
	}

	err = f.ensureDirExists(f.dir(hash))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(f.path(hash), blob, 0644)
}

// PutSD stores the sd blob on the disk
func (f *FileBlobStore) PutSD(hash string, blob []byte) error {
	return f.Put(hash, blob)
}

// Delete deletes the blob from the store
func (f *FileBlobStore) Delete(hash string) error {
	err := f.initOnce()
	if err != nil {
		return err
	}

	has, err := f.Has(hash)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	return os.Remove(f.path(hash))
}
