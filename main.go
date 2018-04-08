// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/swinslow/lfscanning/config"
	"github.com/swinslow/lfscanning/database"
	"github.com/swinslow/lfscanning/repomanager"
)

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
	var id int

	// repo, err = db.InsertRepo("swinslow", "fabric")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// id = repo.Id
	id = 9

	repo, err = db.GetRepoById(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", repo)

	repoPath := rm.GetPathToRepo(repo)
	fmt.Printf("Path to repo: %s\n", repoPath)
	repoURL := rm.GetURLToRepo(repo)
	fmt.Printf("URL to repo:  %s\n", repoURL)

	// err = rm.CloneRepo(repo)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Printf("Cloned %s to %s\n", repoPath, repoURL)
	// }

	fmt.Println("Updating repo...")
	err = rm.UpdateRepo(repo)
	if err != nil {
		fmt.Println(err)
		return
	}

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
