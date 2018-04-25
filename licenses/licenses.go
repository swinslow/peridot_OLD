// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package licenses

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SPDXLicense represents data about a single license from the
// SPDX License List.
type SPDXLicense struct {
	Reference       string `json:"reference"`
	IsDeprecated    bool   `json:"isDeprecatedLicenseId"`
	DetailsURL      string `json:"detailsUrl"`
	ReferenceNumber int    `json:"referenceNumber"`
	Name            string `json:"name"`
	Identifier      string `json:"licenseId"`
	IsOSIApproved   bool   `json:"isOsiApproved"`
}

// SPDXLicenseList represents all SPDX licenses currently on the
// SPDX License List.
type SPDXLicenseList struct {
	LicenseListVersion string        `json:"licenseListVersion"`
	Licenses           []SPDXLicense `json:"licenses"`
	ReleaseDate        string        `json:"releaseDate"`
}

// LoadFromJSON takes a path to an SPDX License List Data JSON directory
// and returns, if successful, an SPDXLicenseList containing its data.
func LoadFromJSON(spdxLLJSONLocation string) (*SPDXLicenseList, error) {
	fpath := filepath.Join(spdxLLJSONLocation, "licenses.json")
	f, err := os.Open(fpath)
	if err != nil {
		return nil, fmt.Errorf("couldn't open license list file: %v", err)
	}
	defer f.Close()

	jsonBytes := make([]byte, 1024)
	var jsonString string
	for {
		n, err := f.Read(jsonBytes)
		if 0 == n {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading license list file: %v", err)
		}
		jsonString += string(jsonBytes[:n])
	}

	var ll SPDXLicenseList
	json.Unmarshal([]byte(jsonString), &ll)
	return &ll, nil
}
