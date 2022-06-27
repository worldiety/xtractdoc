package golang

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PkgDirs returns all directories containing any go file.
func PkgDirs(root string) ([]string, error) {
	var res []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		if d.IsDir() {
			files, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			anyGoFile := false
			for _, file := range files {
				if file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".go") {
					anyGoFile = true
					break
				}
			}

			if anyGoFile {
				res = append(res, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

// ModWdRoot walks up from current working dir to find the enclosing go module.
func ModWdRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working dir: %w", err)
	}

	return ModRoot(cwd)
}

func ModRoot(cwd string) (string, error) {
	for cwd != "." && cwd != "/" && cwd != "" {
		modFile := filepath.Join(cwd, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			return cwd, nil
		}

		cwd = filepath.Dir(cwd)
	}

	return "", fmt.Errorf("no go.mod found")
}

// ModulePath returns whatever the mod path is.
func ModulePath(dir string) (string, error) {
	mdir, err := ModRoot(dir)
	if err != nil {
		return "", err
	}

	mf := filepath.Join(mdir, "go.mod")
	buf, err := os.ReadFile(mf)
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %w", mf, err)
	}

	return modfile.ModulePath(buf), nil
}
