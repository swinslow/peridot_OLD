// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package spdxtvmanager

import (
	"fmt"
	"strings"
)

type filedata struct {
	path               string
	licenseFoundInFile []string
	licenseConcluded   string
	sha1               string
	sha256             string
	md5                string
}

type spdxTVParser struct {
	midfile         bool
	fdList          []*filedata
	currentFileData *filedata
}

func (parser *spdxTVParser) finalize() ([]*filedata, error) {
	if parser.fdList == nil && parser.currentFileData == nil {
		return nil, nil
	}

	// save current record from currentFileData to fdList
	parser.fdList = append(parser.fdList, parser.currentFileData)

	// and reset other values
	parser.midfile = false
	parser.currentFileData = nil

	return parser.fdList, nil
}

func (parser *spdxTVParser) parseNextPair(tag string, value string) error {
	if parser.midfile {
		return parser.parseNextPairFromMidfile(tag, value)
	}

	return parser.parseNextPairFromReady(tag, value)
}

func (parser *spdxTVParser) parseNextPairFromReady(tag string, value string) error {
	switch tag {
	case "FileName":
		parser.currentFileData = &filedata{path: value}
		parser.midfile = true
	}
	return nil
}

func (parser *spdxTVParser) parseNextPairFromMidfile(tag string, value string) error {
	switch tag {
	case "LicenseFoundInFile":
		if parser.currentFileData.licenseFoundInFile == nil {
			parser.currentFileData.licenseFoundInFile = make([]string, 1)
			parser.currentFileData.licenseFoundInFile[0] = value
		} else {
			parser.currentFileData.licenseFoundInFile = append(
				parser.currentFileData.licenseFoundInFile, value)
		}

	case "LicenseConcluded":
		parser.currentFileData.licenseConcluded = value

	case "FileChecksum":
		return parser.parseFileChecksum(value)

	case "FileName":
		// save old record from currentFileData to fdList
		parser.fdList = append(parser.fdList, parser.currentFileData)
		// and start a new one with this path
		parser.currentFileData = &filedata{path: value}
	}

	return nil
}

func (parser *spdxTVParser) parseFileChecksum(checksumTV string) error {
	// parse the FileChecksum value to see if it's a valid checksum type
	sp := strings.SplitN(checksumTV, ":", 2)
	if len(sp) == 1 {
		return fmt.Errorf("invalid FileChecksum format: %s", checksumTV)
	}

	cType := sp[0]
	cSum := sp[1]

	// fail if there's another colon in the checksum section
	colon := strings.SplitN(cSum, ":", 2)
	if len(colon) != 1 {
		return fmt.Errorf("invalid FileChecksum format: %s", checksumTV)
	}

	switch cType {
	case "SHA1":
		parser.currentFileData.sha1 = strings.TrimSpace(cSum)

	case "SHA256":
		parser.currentFileData.sha256 = strings.TrimSpace(cSum)

	case "MD5":
		parser.currentFileData.md5 = strings.TrimSpace(cSum)

	default:
		return fmt.Errorf("unknown FileChecksum type: %s", cType)

	}

	return nil
}
