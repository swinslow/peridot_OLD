// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package config

type Config struct {
	DBConnectString string
	ReposLocation   string
	HashesLocation  string
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
