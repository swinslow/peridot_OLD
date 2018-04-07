// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

type Config struct {
	DBConnectString string
	RepoLocation    string
}

func (cfg *Config) SetDBConnectString(
	user string, password string,
	dbname string, ssl bool) {
	cfg.DBConnectString = ""
	cfg.DBConnectString += "user=" + user
	if password != "" {
		cfg.DBConnectString += " password=" + password
	}
	cfg.DBConnectString += " dbname=" + dbname
	if ssl == true {
		cfg.DBConnectString += " sslmode=verify-full"
	} else {
		cfg.DBConnectString += " sslmode=disable"
	}
}

func (cfg *Config) SetRepoLocation(path string) error {
	// check whether path exists in filesystem
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	// check whether path is a directory
	if !fi.IsDir() {
		return errors.New(path + " is not a directory")
	}

	// check whether path is writable
	if unix.Access(path, unix.W_OK) != nil {
		return errors.New(path + " is not writable")
	}

	// we're good
	cfg.RepoLocation = path
	return nil
}
