// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/swinslow/peridot/cli"
	"github.com/swinslow/peridot/config"
	"github.com/swinslow/peridot/coordinator"
	"github.com/swinslow/peridot/database"
)

func main() {
	var err error

	co := &coordinator.Coordinator{}

	cfg := &config.Config{}
	cfg.SetDBConnectString("steve", "", "peridot", false)
	cfg.ReposLocation = "/Users/steve/programming/scanning/peridot-repos"
	cfg.HashesLocation = "/Users/steve/programming/scanning/peridot-hashes"

	db := database.InitDB()
	err = db.PrepareDB(cfg)
	if err != nil {
		return
	}

	err = co.Prepare(cfg, db)
	if err != nil {
		fmt.Printf("Error preparing coordinator: %v\n", err)
		return
	}

	if len(os.Args) < 2 {
		fmt.Printf("Must specify command; available commands:\n")
		printCommands()
		return
	}

	command := os.Args[1]
	switch command {
	case "repo":
		cli.CmdRepo(co, db, cfg)
	case "reset":
		cli.CmdReset(co, db, cfg)
	default:
		fmt.Printf("Invalid command %s; available commands:\n")
		printCommands()
	}
}

func printCommands() {
	fmt.Printf("  repo\n")
	fmt.Printf("  reset\n")
	fmt.Printf("\n")
}
