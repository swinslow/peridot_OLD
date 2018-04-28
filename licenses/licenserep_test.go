// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package licenses

import (
	"testing"
)

func printNodeExpression(t *testing.T, node *ParsedLicenseNode) {
	if node == nil {
		return
	}
	printNodeExpression(t, node.LeftChild)
	printNodeExpression(t, node.RightChild)
	t.Logf(node.Expression)
}

func TestCanGetNodesForExpression(t *testing.T) {
	expr := "MIT AND (Apache-2.0 OR BSD-2-Clause) AND Zlib"
	//nodeTree, err := GetNodesForExpression(expr)
	_, err := GetNodesForExpression(expr)
	if err != nil {
		t.Fatalf("error from GetNodesForExpression: %v", err)
	}

	//printNodeExpression(t, nodeTree)
}
