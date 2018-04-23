// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"os"

	"github.com/swinslow/peridot/config"
	"github.com/swinslow/peridot/coordinator"
	"github.com/swinslow/peridot/database"
)

type repoCallData struct {
	co       *coordinator.Coordinator
	db       *database.DB
	cfg      *config.Config
	subCmd   string
	orgName  string
	repoName string
}

// CmdRepo provides the "repo" cli command, which is used to initialize or update
// a repo within peridot.
func CmdRepo(co *coordinator.Coordinator, db *database.DB, cfg *config.Config) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: %s repo SUBCOMMAND orgName repoName\n", os.Args[0])
		fmt.Printf("Available subcommands:\n")
		printRepoSubcommands()
		return
	}

	rcd := &repoCallData{co, db, cfg, os.Args[2], os.Args[3], os.Args[4]}

	switch rcd.subCmd {
	case "init":
		subcmdRepoInit(rcd)
	case "update":
		subcmdRepoUpdate(rcd)
	default:
		printRepoSubcommands()
	}
}

func printRepoSubcommands() {
	fmt.Printf("  init\n")
	fmt.Printf("  update\n")
	fmt.Printf("  info   (NOT YET IMPLEMENTED)\n")
	fmt.Printf("  delete (NOT YET IMPLEMENTED)\n")
}

func subcmdRepoInit(rcd *repoCallData) {
	var err error
	var repoID int

	// first make sure that the repo isn't already in the database
	repoID, err = rcd.db.GetRepoIDFromCoords(rcd.orgName, rcd.repoName)
	if err != nil {
		fmt.Printf("Error getting repo ID: %v\n", err)
		return
	}
	if repoID != 0 {
		fmt.Printf("Error in 'repo init': %s/%s already exists in database\n", rcd.orgName, rcd.repoName)
		fmt.Printf("Did you mean to call 'repo update' instead?\n")
		return
	}

	// repo doesn't yet exist in database, so set it up and get an ID
	repo, err := rcd.db.InsertRepo(rcd.orgName, rcd.repoName)
	if err != nil {
		fmt.Printf("Error adding repo to DB: %v\n", err)
		return
	}
	repoID = repo.Id

	// go clone the repo from remote
	fmt.Printf("Getting Github repo %s/%s...\n", rcd.orgName, rcd.repoName)
	err = rcd.co.DoCloneRepo(repo.Id)
	if err != nil {
		fmt.Printf("Error cloning repo: %v\n", err)
		return
	}
	fmt.Printf("Cloned repo\n")

	// and prepare the directories and files in the database
	err = rcd.co.DoPrepareFiles(repoID)
	if err != nil {
		fmt.Printf("Error preparing files: %v\n", err)
		return
	}
}

func subcmdRepoUpdate(rcd *repoCallData) {
	var err error
	var repoID int
	var needsFilesPrepared bool

	// first make sure that the repo is already in the database
	repoID, err = rcd.db.GetRepoIDFromCoords(rcd.orgName, rcd.repoName)
	if err != nil {
		fmt.Printf("Error getting repo ID: %v\n", err)
		return
	}
	if repoID == 0 {
		fmt.Printf("Error in 'repo update': %s/%s not found in database\n", rcd.orgName, rcd.repoName)
		fmt.Printf("Did you mean to call 'repo init' instead?\n")
		return
	}

	// repo already exists, so let's update it
	fmt.Printf("Checking Github repo %s/%s for updates...\n", rcd.orgName, rcd.repoName)
	needsFilesPrepared, err = rcd.co.DoUpdateRepo(repoID)
	if err != nil {
		fmt.Printf("Error updating repo: %v\n", err)
		return
	}
	fmt.Printf("Checked for updates\n")

	if needsFilesPrepared {
		fmt.Printf("Updates found, so preparing directories and files\n")
		err = rcd.co.DoPrepareFiles(repoID)
		if err != nil {
			fmt.Printf("Error preparing files: %v\n", err)
			return
		}
	} else {
		fmt.Printf("No updates found\n")
	}
}
