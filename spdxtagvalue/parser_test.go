// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package spdxtagvalue

import "testing"

func TestFilenameTagInReadyMovesToMidfile(t *testing.T) {
	parser := &spdxTVParser{}
	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	if parser.midfile == false {
		t.Errorf("expected midfile to be true, got false")
	}
	if parser.currentFileData == nil {
		t.Errorf("currentFileData is nil after calling parseNextPair")
	}
	if parser.currentFileData.path != "/tmp/hi" {
		t.Errorf("expected %s for currentFileData.path, got %s", "/tmp/hi", parser.currentFileData.path)
	}
	if parser.currentFileData.licenseConcluded != "" {
		t.Errorf("expected empty string for currentFileData.licenseConcluded, got %s", parser.currentFileData.licenseConcluded)
	}
	if parser.currentFileData.licenseFoundInFile != nil {
		t.Errorf("expected empty string for currentFileData.licenseFoundInFile, got %s", parser.currentFileData.licenseFoundInFile)
	}
	if parser.currentFileData.sha1 != "" {
		t.Errorf("expected empty string for currentFileData.sha1, got %s", parser.currentFileData.sha1)
	}
	if parser.currentFileData.sha256 != "" {
		t.Errorf("expected empty string for currentFileData.sha256, got %s", parser.currentFileData.sha256)
	}
	if parser.currentFileData.md5 != "" {
		t.Errorf("expected empty string for currentFileData.md5, got %s", parser.currentFileData.md5)
	}
	if len(parser.fdList) != 0 {
		t.Errorf("expected fdList to be empty, got %v", parser.fdList)
	}
}

func TestNonFilenameTagInReadyStaysInReady(t *testing.T) {
	parser := &spdxTVParser{}
	err := parser.parseNextPair("whatever", "something else")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	if parser.midfile == true {
		t.Errorf("expected midfile to be false, got true")
	}
	if parser.currentFileData != nil {
		t.Errorf("expected currentFileData to be nil after calling parseNextPair, got %v", parser.currentFileData)
	}
}

func TestCanExtractChecksumTypes(t *testing.T) {
	parser := &spdxTVParser{}
	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseFileChecksum("SHA1: abc123")
	if err != nil {
		t.Errorf("got error when calling parseFileChecksum: %v", err)
	}
	if parser.currentFileData == nil {
		t.Errorf("currentFileData is nil after calling parseFileChecksum")
	}
	if parser.currentFileData.sha1 != "abc123" {
		t.Errorf("expected %s for currentFileData.sha1, got %s", "abc123", parser.currentFileData.sha1)
	}

	err = parser.parseFileChecksum("SHA256: def432")
	if err != nil {
		t.Errorf("got error when calling parseFileChecksum: %v", err)
	}
	if parser.currentFileData == nil {
		t.Errorf("currentFileData is nil after calling parseFileChecksum")
	}
	if parser.currentFileData.sha1 != "abc123" {
		t.Errorf("expected %s for currentFileData.sha1, got %s", "abc123", parser.currentFileData.sha1)
	}
	if parser.currentFileData.sha256 != "def432" {
		t.Errorf("expected %s for currentFileData.sha256, got %s", "def432", parser.currentFileData.sha1)
	}

	err = parser.parseFileChecksum("MD5: 035183")
	if err != nil {
		t.Errorf("got error when calling parseFileChecksum: %v", err)
	}
	if parser.currentFileData == nil {
		t.Errorf("currentFileData is nil after calling parseFileChecksum")
	}
	if parser.currentFileData.sha1 != "abc123" {
		t.Errorf("expected %s for currentFileData.sha1, got %s", "abc123", parser.currentFileData.sha1)
	}
	if parser.currentFileData.sha256 != "def432" {
		t.Errorf("expected %s for currentFileData.sha256, got %s", "def432", parser.currentFileData.sha1)
	}
	if parser.currentFileData.md5 != "035183" {
		t.Errorf("expected %s for currentFileData.md5, got %s", "035183", parser.currentFileData.sha1)
	}

}

func TestReturnsErrorForInvalidShortChecksumFormat(t *testing.T) {
	parser := &spdxTVParser{}
	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseFileChecksum("blah")
	if err == nil {
		t.Errorf("expected error when calling parseFileChecksum for invalid format, got nil")
	}
}

func TestReturnsErrorForInvalidLongChecksumFormat(t *testing.T) {
	parser := &spdxTVParser{}
	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseFileChecksum("MD5: 12390834: other")
	if err == nil {
		t.Errorf("expected error when calling parseFileChecksum for invalid format, got nil")
	}
}

func TestReturnsErrorForUnknownChecksumType(t *testing.T) {
	parser := &spdxTVParser{}
	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseFileChecksum("ECDSA: 12390834")
	if err == nil {
		t.Errorf("expected error when calling parseFileChecksum for unknown checksum type, got nil")
	}
}

func TestRecordsDataInCurrentFileData(t *testing.T) {
	parser := &spdxTVParser{}
	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseFoundInFile", "MIT")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseConcluded", "Apache-2.0 AND (MIT OR BSD-2-Clause)")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseFoundInFile", "Apache-2.0")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("FileChecksum", "SHA1: abc123")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "SHA256: 456789")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "MD5: 0def12")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	if parser.midfile == false {
		t.Errorf("expected midfile to be true, got false")
	}
	if parser.currentFileData == nil {
		t.Errorf("currentFileData is nil after calls to parseNextPair")
	}
	if parser.currentFileData.path != "/tmp/hi" {
		t.Errorf("expected %s for currentFileData.path, got %s", "/tmp/hi", parser.currentFileData.path)
	}
	if parser.currentFileData.licenseConcluded != "Apache-2.0 AND (MIT OR BSD-2-Clause)" {
		t.Errorf("expected Apache-2.0 AND (MIT OR BSD-2-Clause) for currentFileData.licenseConcluded, got %s", parser.currentFileData.licenseConcluded)
	}
	if parser.currentFileData.licenseFoundInFile == nil ||
		len(parser.currentFileData.licenseFoundInFile) != 2 ||
		parser.currentFileData.licenseFoundInFile[0] != "MIT" ||
		parser.currentFileData.licenseFoundInFile[1] != "Apache-2.0" {
		t.Errorf("expected [MIT Apache-2.0] for currentFileData.licenseFoundInFile, got %s", parser.currentFileData.licenseFoundInFile)
	}
	if parser.currentFileData.sha1 != "abc123" {
		t.Errorf("expected abc123 for currentFileData.sha1, got %s", parser.currentFileData.sha1)
	}
	if parser.currentFileData.sha256 != "456789" {
		t.Errorf("expected 456789 for currentFileData.sha256, got %s", parser.currentFileData.sha256)
	}
	if parser.currentFileData.md5 != "0def12" {
		t.Errorf("expected 0def12 for currentFileData.md5, got %s", parser.currentFileData.md5)
	}
	if len(parser.fdList) != 0 {
		t.Errorf("expected fdList to be empty, got %v", parser.fdList)
	}

}

func TestSavesAndProceedsToNextFileData(t *testing.T) {
	parser := &spdxTVParser{}

	// ===== FIRST FileName =====
	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseFoundInFile", "MIT")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseConcluded", "Apache-2.0 AND (MIT OR BSD-2-Clause)")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseFoundInFile", "Apache-2.0")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("FileChecksum", "SHA1: abc123")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "SHA256: 456789")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "MD5: 0def12")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	// ===== SECOND FileName =====
	err = parser.parseNextPair("FileName", "/tmp/bye")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseFoundInFile", "LicenseRef-Whatever")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseConcluded", "LicenseRef-Whatever")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("FileChecksum", "SHA1: abc")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "SHA256: def")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "MD5:123")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	// ===== Check that FIRST file data got saved in fdList

	if parser.midfile == false {
		t.Errorf("expected midfile to be true, got false")
	}
	if len(parser.fdList) != 1 {
		t.Errorf("expected len(fdList) to be 1, got %d", len(parser.fdList))
	}
	if parser.fdList[0].path != "/tmp/hi" {
		t.Errorf("expected %s for fdList[0].path, got %s", "/tmp/hi", parser.fdList[0].path)
	}
	if parser.fdList[0].licenseConcluded != "Apache-2.0 AND (MIT OR BSD-2-Clause)" {
		t.Errorf("expected Apache-2.0 AND (MIT OR BSD-2-Clause) for fdList[0].licenseConcluded, got %s", parser.fdList[0].licenseConcluded)
	}
	if parser.fdList[0].licenseFoundInFile == nil ||
		len(parser.fdList[0].licenseFoundInFile) != 2 ||
		parser.fdList[0].licenseFoundInFile[0] != "MIT" ||
		parser.fdList[0].licenseFoundInFile[1] != "Apache-2.0" {
		t.Errorf("expected [MIT Apache-2.0] for fdList[0].licenseFoundInFile, got %s", parser.fdList[0].licenseFoundInFile)
	}
	if parser.fdList[0].sha1 != "abc123" {
		t.Errorf("expected abc123 for fdList[0].sha1, got %s", parser.fdList[0].sha1)
	}
	if parser.fdList[0].sha256 != "456789" {
		t.Errorf("expected 456789 for fdList[0].sha256, got %s", parser.fdList[0].sha256)
	}
	if parser.fdList[0].md5 != "0def12" {
		t.Errorf("expected 0def12 for fdList[0].md5, got %s", parser.fdList[0].md5)
	}

	// ===== Check that SECOND file data is stored in currentFileData

	if parser.currentFileData == nil {
		t.Errorf("currentFileData is nil after calls to parseNextPair")
	}
	if parser.currentFileData.path != "/tmp/bye" {
		t.Errorf("expected %s for currentFileData.path, got %s", "/tmp/bye", parser.currentFileData.path)
	}
	if parser.currentFileData.licenseConcluded != "LicenseRef-Whatever" {
		t.Errorf("expected LicenseRef-Whatever for currentFileData.licenseConcluded, got %s", parser.currentFileData.licenseConcluded)
	}
	if parser.currentFileData.licenseFoundInFile == nil ||
		len(parser.currentFileData.licenseFoundInFile) != 1 ||
		parser.currentFileData.licenseFoundInFile[0] != "LicenseRef-Whatever" {
		t.Errorf("expected [LicenseRef-Whatever] for currentFileData.licenseFoundInFile, got %s", parser.currentFileData.licenseFoundInFile)
	}
	if parser.currentFileData.sha1 != "abc" {
		t.Errorf("expected abc for currentFileData.sha1, got %s", parser.currentFileData.sha1)
	}
	if parser.currentFileData.sha256 != "def" {
		t.Errorf("expected def for currentFileData.sha256, got %s", parser.currentFileData.sha256)
	}
	if parser.currentFileData.md5 != "123" {
		t.Errorf("expected 123 for currentFileData.md5, got %s", parser.currentFileData.md5)
	}

}

func TestFinalizeSavesLastFileData(t *testing.T) {
	parser := &spdxTVParser{}

	err := parser.parseNextPair("FileName", "/tmp/hi")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseFoundInFile", "MIT")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseConcluded", "Apache-2.0 AND (MIT OR BSD-2-Clause)")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("LicenseFoundInFile", "Apache-2.0")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	err = parser.parseNextPair("FileChecksum", "SHA1: abc123")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "SHA256: 456789")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}
	err = parser.parseNextPair("FileChecksum", "MD5: 0def12")
	if err != nil {
		t.Errorf("got error when calling parseNextPair: %v", err)
	}

	fdList, err := parser.finalize()
	if err != nil {
		t.Errorf("got error when calling finalize: %v", err)
	}

	if parser.midfile == true {
		t.Errorf("expected midfile to be false, got true")
	}
	if len(fdList) != 1 {
		t.Errorf("expected len(fdList) to be 1, got %d", len(fdList))
	}
	if fdList[0].path != "/tmp/hi" {
		t.Errorf("expected %s for fdList[0].path, got %s", "/tmp/hi", fdList[0].path)
	}
	if fdList[0].licenseConcluded != "Apache-2.0 AND (MIT OR BSD-2-Clause)" {
		t.Errorf("expected Apache-2.0 AND (MIT OR BSD-2-Clause) for fdList[0].licenseConcluded, got %s", fdList[0].licenseConcluded)
	}
	if fdList[0].licenseFoundInFile == nil ||
		len(fdList[0].licenseFoundInFile) != 2 ||
		fdList[0].licenseFoundInFile[0] != "MIT" ||
		fdList[0].licenseFoundInFile[1] != "Apache-2.0" {
		t.Errorf("expected [MIT Apache-2.0] for fdList[0].licenseFoundInFile, got %s", fdList[0].licenseFoundInFile)
	}
	if fdList[0].sha1 != "abc123" {
		t.Errorf("expected abc123 for fdList[0].sha1, got %s", fdList[0].sha1)
	}
	if fdList[0].sha256 != "456789" {
		t.Errorf("expected 456789 for fdList[0].sha256, got %s", fdList[0].sha256)
	}
	if fdList[0].md5 != "0def12" {
		t.Errorf("expected 0def12 for fdList[0].md5, got %s", fdList[0].md5)
	}

	if parser.currentFileData != nil {
		t.Errorf("expected currentFileData to be nil, got %v", parser.currentFileData)
	}

}
