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

	repo, err = addNewRepo(db, "swinslow", "fabric")
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

	curRetrievalId := 0
	repoId := repo.Id

	// ===== OPTION 2: retrieve existing repo and update it
	/*
		repoId := 1

		repo, err = db.GetRepoById(repoId)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%#v\n", repo)

		curRetrieval, err := db.GetRepoRetrievalLatest(repoId)
		if err != nil {
			fmt.Printf("Error getting repo retrieval before update: %v\n", err)
			return
		}
		curRetrievalId := curRetrieval.Id
	*/

	// ===== Update repo

	fmt.Println("Updating repo...")
	err = rm.UpdateRepo(repo)
	if err != nil {
		fmt.Println(err)
		return
	}
	repoRetrieval, err := db.GetRepoRetrievalLatest(repoId)
	if err != nil {
		fmt.Printf("Error getting repo retrieval: %v\n", err)
		return
	}

	// ========== Now, do whatever actions we want

	// msg, err := rm.GetRepoLatestCommit(repo)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("Latest commit message: %s\n", msg)

	//rm.WalkAndPrintFiles(repo, "slm/commands/cmdEditCategory.py")

	if repoRetrieval.Id != curRetrievalId {
		// we updated with a new retrieval; re-scan
		allPaths, err := rm.GetAllFilepaths(repo)
		if err != nil {
			fmt.Println(err)
			return
		}

		dirPaths := database.ExtractDirsFromPaths(allPaths)
		// for i, dirPath := range dirPaths {
		// 	fmt.Printf("%d: %s\n", i, dirPath)
		// }

		err = db.BulkInsertRepoDirs(repoRetrieval.Id, dirPaths)
		if err != nil {
			fmt.Printf("Error inserting repo directories into database: %v\n", err)
			return
		}

		pathsToHashes, err := rm.GetFileHashes(repo)
		if err != nil {
			fmt.Printf("Error getting file hashes: %v\n", err)
			return
		}

		err = db.BulkInsertRepoFiles(repoRetrieval.Id, pathsToHashes)
		if err != nil {
			fmt.Printf("Error inserting repo files into database: %v\n", err)
			return
		}

		repoFiles, err := db.GetRepoFilesForRepoRetrieval(repoRetrieval.Id)
		if err != nil {
			fmt.Printf("Error getting repo files: %v\n", err)
			return
		}
		for _, repoFile := range repoFiles {
			fmt.Printf("%d (%s): PREV %d (%s), NEXT %d (%s)\n",
				repoFile.Id, repoFile.Path,
				repoFile.PrevFileId, repoFiles[repoFile.PrevFileId].Path,
				repoFile.NextFileId, repoFiles[repoFile.NextFileId].Path,
			)
		}
	}
}
