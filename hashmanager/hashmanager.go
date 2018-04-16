// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package hashmanager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/swinslow/peridot/config"
	"github.com/swinslow/peridot/database"
)

type HashManager struct {
	HashesPath string
	db         *database.DB
}

func (hm *HashManager) PrepareHM(cfg *config.Config, db *database.DB) error {
	if hm == nil {
		return fmt.Errorf("must pass non-nil HashManager")
	}
	if cfg == nil || cfg.ReposLocation == "" {
		return fmt.Errorf("must pass config string")
	}
	if db == nil {
		return fmt.Errorf("must prepare and pass database")
	}

	err := hm.setHashesLocation(cfg.HashesLocation)
	if err != nil {
		return err
	}

	hm.db = db
	return nil
}

func (hm *HashManager) setHashesLocation(path string) error {
	// check whether path exists in filesystem
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	// check whether path is a directory
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	// check whether path is writable
	if unix.Access(path, unix.W_OK) != nil {
		return fmt.Errorf("%s is not writable", path)
	}

	// we're good
	hm.HashesPath = path
	return nil
}

func (hm *HashManager) GetPathToHash(h_sha1 string, h_sha256 string, h_md5 string) string {
	// uses all three hashes for dirs
	// uses just SHA1 and SHA256 for filename
	return filepath.Join(hm.HashesPath, h_sha1[:2], h_sha256[:2], h_md5[:2],
		fmt.Sprintf("%s.%s", h_sha1, h_sha256))
}

// returns false, nil if file already exists in the hash location
func (hm *HashManager) CopyFileToHash(srcPath string, h_sha1 string, h_sha256 string, h_md5 string) (bool, error) {
	// first check if there's already a file in the dst path
	dstPath := hm.GetPathToHash(h_sha1, h_sha256, h_md5)
	_, err := os.Stat(dstPath)
	if err == nil {
		// no error from stat means a file exists at dstPath
		return false, nil
	}

	if !(os.IsNotExist(err)) {
		return false, fmt.Errorf("couldn't check hash location: %v", err)
	}

	// dstPath is available
	// make sure the directory exists
	err = os.MkdirAll(filepath.Dir(dstPath), 0700)
	if err != nil {
		return false, fmt.Errorf("couldn't create hash subdir: %v", err)
	}

	// and copy the file there
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return false, fmt.Errorf("couldn't open src file for copying: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return false, fmt.Errorf("couldn't open dst file for copying: %v", err)
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		dstFile.Close()
		return false, fmt.Errorf("error copying file: %v", err)
	}

	err = dstFile.Close()
	if err != nil {
		return false, fmt.Errorf("error closing file: %v", err)
	}

	return true, nil
}

// returns new map of path => array of hashes for just the newly-copied files
func (hm *HashManager) CopyAllFilesToHash(pathRoot string, pathsToHashes map[string][3]string) (map[string][3]string, error) {
	copiedFiles := make(map[string][3]string)

	for path, hashes := range pathsToHashes {
		fullPath := filepath.Join(pathRoot, path)
		copied, err := hm.CopyFileToHash(fullPath, hashes[0], hashes[1], hashes[2])
		if err != nil {
			return nil, fmt.Errorf("error copying all files to hashes: %v", err)
		}

		if copied {
			copiedFiles[path] = hashes
		}
	}
	return copiedFiles, nil
}