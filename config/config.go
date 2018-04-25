// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package config

// Config represents data for configuring peridot.
type Config struct {
	DBConnectString    string
	ReposLocation      string
	HashesLocation     string
	SPDXLLJSONLocation string
}

// SetDBConnectString is called with database config paramters to create the
// appropriate connection string in its Config object.
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
