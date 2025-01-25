package fs

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

func MkdirAllWithChown(fs afero.Fs, path string, perm os.FileMode, uid int, gid int) (err error) {
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	pathList := strings.Split(path, string(filepath.Separator))
	pathList[0] = string(filepath.Separator)
	for i := 1; i < len(pathList); i++ {
		dir := filepath.Join(pathList[:i+1]...)
		if ok, _ := afero.Exists(fs, dir); ok {
			continue
		}

		err = fs.Mkdir(dir, perm)
		if err != nil {
			return err
		}

		err = fs.Chown(dir, uid, gid)
		if err != nil {
			return err
		}
	}

	return nil
}
