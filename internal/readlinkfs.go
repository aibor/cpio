package internal

import (
	"errors"
	gofs "io/fs"
	"os"
)

// Replace with [gofs.ReadLinkFS] once available (planned for 1.25). See
// https://github.com/golang/go/issues/49580
type ReadLinkFS interface {
	gofs.FS

	ReadLink(name string) (string, error)
	Lstat(name string) (gofs.FileInfo, error)
}

func ReadLink(fsys gofs.FS, name string) (string, error) {
	rlFS, ok := fsys.(ReadLinkFS)
	if !ok {
		return "", &gofs.PathError{
			Op:   "readlink",
			Path: name,
			Err:  gofs.ErrInvalid,
		}
	}

	return rlFS.ReadLink(name)
}

var _ ReadLinkFS = (*readLinkFS)(nil)

type readLinkFS struct {
	gofs.FS
	readLinkFn func(name string) (string, error)
	lstatFn    func(name string) (gofs.FileInfo, error)
}

func (fsys *readLinkFS) ReadLink(name string) (string, error) {
	return fsys.readLinkFn(name)
}

func (fsys *readLinkFS) Lstat(name string) (gofs.FileInfo, error) {
	return fsys.lstatFn(name)
}

// Replace with [os.DirFS] once available (planned for 1.25). See
// https://github.com/golang/go/issues/49580
func ReadLinkDirFS(dir string) gofs.FS {
	join := func(name string) (string, error) {
		if dir == "" {
			return "", errors.New("os: DirFS with empty root")
		}
		if !gofs.ValidPath(name) {
			return "", os.ErrInvalid
		}
		if os.IsPathSeparator(dir[len(dir)-1]) {
			return string(dir) + name, nil
		}
		return string(dir) + string(os.PathSeparator) + name, nil
	}

	return &readLinkFS{
		FS: os.DirFS(dir),
		readLinkFn: func(name string) (string, error) {
			path, err := join(name)
			if err != nil {
				return "", err
			}
			return os.Readlink(path)
		},
		lstatFn: func(name string) (gofs.FileInfo, error) {
			path, err := join(name)
			if err != nil {
				return nil, err
			}
			return os.Lstat(path)
		},
	}
}
