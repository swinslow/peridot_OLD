// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/swinslow/lfscanning/config"
	"github.com/swinslow/lfscanning/database"
)

func main() {
	cfg := &config.Config{}
	cfg.SetDBConnectString("steve", "", "lfscanning", false)

	var db database.DB
	err := db.PrepareDB(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = db.CreateDBTablesIfNotExists()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cfg.SetRepoLocation("/home/steve/programming/lftools/lfscanning/repos")
	fmt.Println(err)

	repo, err := db.GetRepoById(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", repo)
}
