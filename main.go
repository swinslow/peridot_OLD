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

	db, err := database.PrepareDB(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = database.CreateDBTablesIfNotExists(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cfg.SetRepoLocation("/home/steve/programming/lftools/lfscanning/repos")
	fmt.Println(err)

	repo, err := database.GetRepoById(db, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", repo)
}
