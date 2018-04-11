// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/swinslow/lfscanning/config"
	"github.com/swinslow/lfscanning/database"
	"github.com/swinslow/lfscanning/repomanager"
)

func addNewRepo(db *database.DB, orgName string, repoName string) (*database.Repo, error) {
	repo, err := db.InsertRepo(orgName, repoName)
	if err != nil {
		return nil, err
	}
	return repo, err
}

func main() {
	var err error

	cfg := &config.Config{}
	cfg.SetDBConnectString("steve", "", "lfscanning", false)
	cfg.ReposLocation = "/home/steve/programming/lftools/lfscanning/repos"

	db := database.InitDB()
	err = db.PrepareDB(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	rm := repomanager.InitRepoManager()
	err = rm.PrepareRepoManager(cfg, db)
	if err != nil {
		fmt.Println(err)
		return
	}

	var repo *database.Repo

	// ===== OPTION 1: add new repo
	/*
		repo, err = addNewRepo(db, "swinslow", "testrepo")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = rm.CloneRepo(repo)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Cloned %s to %s\n", rm.GetPathToRepo(repo), rm.GetURLToRepo(repo))
		}
	*/
	//id := repo.Id

	// ===== OPTION 2: retrieve existing repo and update it
	id := 1
	repo, err = db.GetRepoById(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", repo)
	fmt.Println("Updating repo...")
	err = rm.UpdateRepo(repo)
	if err != nil {
		fmt.Println(err)
		return
	}

	// ========== Now, do whatever actions we want

	msg, err := rm.GetRepoLatestCommit(repo)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Latest commit message: %s\n", msg)

	//rm.WalkAndPrintFiles(repo, "slm/commands/cmdEditCategory.py")

	allPaths, err := rm.GetAllFilepaths(repo)
	if err != nil {
		fmt.Println(err)
		return
	}

	dirPaths := database.ExtractDirsFromPaths(allPaths)
	for i, dirPath := range dirPaths {
		fmt.Printf("%d: %s\n", i, dirPath)
	}
}
