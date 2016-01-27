package security

import (
	"fmt"
	"path/filepath"
)

// FIXME: duplicated code

func originFromBasedir(basedir string) (string, error) {
	ext := filepath.Ext(filepath.Dir(filepath.Clean(basedir)))
	if len(ext) < 2 {
		return "", fmt.Errorf("can not get origin from path %q", basedir)
	}

	return ext[1:], nil
}
