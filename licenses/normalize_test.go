// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package licenses

import (
	"testing"
)

func TestSimpleNormalization(t *testing.T) {
	var expr string
	var exprTokens []exprToken
	var err error

	expr = "MIT"
	exprTokens, err = getTokens(expr)
	if err != nil {
		t.Errorf(`getTokens("%s") got error: %v`, expr, err)
	}
	if len(exprTokens) != 1 {
		t.Errorf(`getTokens("%s") length = %d`, expr, len(exprTokens))
	}
	if exprTokens[0].t != "IDENTIFIER" || exprTokens[0].identifier != expr {
		t.Errorf(`getTokens("%s")[0] = %v`, expr, exprTokens[0])
	}

	expr = "Apache-2.0"
	exprTokens, err = getTokens(expr)
	if err != nil {
		t.Errorf(`getTokens("%s") got error: %v`, expr, err)
	}
	if len(exprTokens) != 1 {
		t.Errorf(`getTokens("%s") length = %d`, expr, len(exprTokens))
	}
	if exprTokens[0].t != "IDENTIFIER" || exprTokens[0].identifier != expr {
		t.Errorf(`getTokens("%s")[0] = %v`, expr, exprTokens[0])
	}
}

func TestNormalizeWithConjunctions(t *testing.T) {
	var expr string
	var exprTokens []exprToken
	var err error

	expr = "MIT AND BSD-2-Clause"
	exprTokens, err = getTokens(expr)
	if err != nil {
		t.Errorf(`getTokens("%s") got error: %v`, expr, err)
	}
	if len(exprTokens) != 3 {
		t.Errorf(`getTokens("%s") length = %d`, expr, len(exprTokens))
	}
	if exprTokens[0].t != "IDENTIFIER" || exprTokens[0].identifier != "MIT" {
		t.Errorf(`getTokens("%s")[0] = %v`, expr, exprTokens[0])
	}
	if exprTokens[1].t != "AND" || exprTokens[1].identifier != "AND" {
		t.Errorf(`getTokens("%s")[1] = %v`, expr, exprTokens[1])
	}
	if exprTokens[2].t != "IDENTIFIER" || exprTokens[2].identifier != "BSD-2-Clause" {
		t.Errorf(`getTokens("%s")[2] = %v`, expr, exprTokens[2])
	}

	expr = "MIT OR BSD-2-Clause"
	exprTokens, err = getTokens(expr)
	if err != nil {
		t.Errorf(`getTokens("%s") got error: %v`, expr, err)
	}
	if len(exprTokens) != 3 {
		t.Errorf(`getTokens("%s") length = %d`, expr, len(exprTokens))
	}
	if exprTokens[0].t != "IDENTIFIER" || exprTokens[0].identifier != "MIT" {
		t.Errorf(`getTokens("%s")[0] = %v`, expr, exprTokens[0])
	}
	if exprTokens[1].t != "OR" || exprTokens[1].identifier != "OR" {
		t.Errorf(`getTokens("%s")[1] = %v`, expr, exprTokens[1])
	}
	if exprTokens[2].t != "IDENTIFIER" || exprTokens[2].identifier != "BSD-2-Clause" {
		t.Errorf(`getTokens("%s")[2] = %v`, expr, exprTokens[2])
	}

	expr = "MPL-2.0+"
	exprTokens, err = getTokens(expr)
	if err != nil {
		t.Errorf(`getTokens("%s") got error: %v`, expr, err)
	}
	if len(exprTokens) != 2 {
		t.Errorf(`getTokens("%s") length = %d`, expr, len(exprTokens))
	}
	if exprTokens[0].t != "IDENTIFIER" || exprTokens[0].identifier != "MPL-2.0" {
		t.Errorf(`getTokens("%s")[0] = %v`, expr, exprTokens[0])
	}
	if exprTokens[1].t != "PLUS" || exprTokens[1].identifier != "+" {
		t.Errorf(`getTokens("%s")[1] = %v`, expr, exprTokens[1])
	}

	expr = "MPL-2.0+ AND (BSD-2-Clause OR GPL-2.0-or-later)"
	exprTokens, err = getTokens(expr)
	if err != nil {
		t.Errorf(`getTokens("%s") got error: %v`, expr, err)
	}
	if len(exprTokens) != 8 {
		t.Errorf(`getTokens("%s") length = %d`, expr, len(exprTokens))
	}
	if exprTokens[0].t != "IDENTIFIER" || exprTokens[0].identifier != "MPL-2.0" {
		t.Errorf(`getTokens("%s")[0] = %v`, expr, exprTokens[0])
	}
	if exprTokens[1].t != "PLUS" || exprTokens[1].identifier != "+" {
		t.Errorf(`getTokens("%s")[1] = %v`, expr, exprTokens[1])
	}
	if exprTokens[2].t != "AND" || exprTokens[2].identifier != "AND" {
		t.Errorf(`getTokens("%s")[2] = %v`, expr, exprTokens[2])
	}
	if exprTokens[3].t != "LPAREN" || exprTokens[3].identifier != "(" {
		t.Errorf(`getTokens("%s")[3] = %v`, expr, exprTokens[3])
	}
	if exprTokens[4].t != "IDENTIFIER" || exprTokens[4].identifier != "BSD-2-Clause" {
		t.Errorf(`getTokens("%s")[4] = %v`, expr, exprTokens[4])
	}
	if exprTokens[5].t != "OR" || exprTokens[5].identifier != "OR" {
		t.Errorf(`getTokens("%s")[5] = %v`, expr, exprTokens[5])
	}
	if exprTokens[6].t != "IDENTIFIER" || exprTokens[6].identifier != "GPL-2.0-or-later" {
		t.Errorf(`getTokens("%s")[6] = %v`, expr, exprTokens[6])
	}
	if exprTokens[7].t != "RPAREN" || exprTokens[7].identifier != ")" {
		t.Errorf(`getTokens("%s")[7] = %v`, expr, exprTokens[7])
	}

}

func TestConvertNodeTreeToStrings(t *testing.T) {
	var lNode *licenseNode
	var expr string
	var err error

	lNode = &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"}
	expr, err = getExpressionFromNodeTree(lNode)
	if err != nil || expr != "MIT" {
		t.Errorf(`getExpressionFromNodeTree(%v) = %s`, lNode, expr)
		t.Errorf(`plain identifier should be brought down directly`)
	}

	lNode = &licenseNode{nodeType: "IDENTIFIER", identifier: "EPL-1.0", plus: true}
	expr, err = getExpressionFromNodeTree(lNode)
	if err != nil || expr != "EPL-1.0+" {
		t.Errorf(`getExpressionFromNodeTree(%v) = %s`, lNode, expr)
		t.Errorf(`+ suffix should be added when plain identifier has plus=true`)
	}

	lNode = &licenseNode{nodeType: "AND"}
	lNode.leftChild = &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"}
	lNode.rightChild = &licenseNode{nodeType: "IDENTIFIER", identifier: "Apache-2.0"}
	// NOTE that this is intentionally not normalized since it is only testing the
	// string output functionality. The licenseNode tree is not in normalized order;
	// Apache-2.0 should be the leftChild and MIT the rightChild.
	expr, err = getExpressionFromNodeTree(lNode)
	if err != nil || expr != "MIT AND Apache-2.0" {
		t.Errorf(`getExpressionFromNodeTree(%v) = %s`, lNode, expr)
	}

	lNode = &licenseNode{nodeType: "AND"}
	lNode.leftChild = &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"}
	lNode.rightChild = &licenseNode{nodeType: "OR"}
	lNode.rightChild.leftChild = &licenseNode{nodeType: "IDENTIFIER", identifier: "BSD-2-Clause"}
	lNode.rightChild.rightChild = &licenseNode{nodeType: "IDENTIFIER", identifier: "GPL-2.0-or-later"}
	// NOTE that this is intentionally not normalized
	expr, err = getExpressionFromNodeTree(lNode)
	if err != nil || expr != "MIT AND (BSD-2-Clause OR GPL-2.0-or-later)" {
		t.Errorf(`getExpressionFromNodeTree(%v) = %s`, lNode, expr)
	}

}

func compareTrees(t1 *licenseNode, t2 *licenseNode) bool {
	if t1 == nil && t2 == nil {
		return true
	}
	if t1 == nil || t2 == nil {
		// one, but not both, are nil, so fail
		return false
	}
	return t1.nodeType == t2.nodeType &&
		t1.identifier == t2.identifier &&
		t1.plus == t2.plus &&
		compareTrees(t1.leftChild, t2.leftChild) &&
		compareTrees(t1.rightChild, t2.rightChild)
}

func parseAndCompare(t *testing.T, expr string, t2 *licenseNode) {
	exprTokens, err := getTokens(expr)
	if err != nil {
		t.Errorf("error getting tokens for %s: %v", expr, err)
		return
	}

	t1, err := parseTokens(exprTokens)
	if err != nil {
		t.Errorf("error parsing tokens for %v: %v", exprTokens, err)
		return
	}

	if !compareTrees(t1, t2) {
		t.Errorf("tree for %s (%v) does not match test tree (%v)", expr,
			getNodeTreeString(t1), getNodeTreeString(t2))
		return
	}
}

func TestConvertTokensToNodeTree(t *testing.T) {
	var nodeTree *licenseNode

	// single identifier
	nodeTree = &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"}
	parseAndCompare(t, "MIT", nodeTree)

	// parens around single identifier
	nodeTree = &licenseNode{nodeType: "IDENTIFIER", identifier: "BSD-2-Clause"}
	parseAndCompare(t, "(BSD-2-Clause)", nodeTree)

	// single identifier with plus
	nodeTree = &licenseNode{nodeType: "IDENTIFIER", identifier: "MPL-2.0", plus: true}
	parseAndCompare(t, "MPL-2.0+", nodeTree)

	// basic complex expression
	nodeTree = &licenseNode{nodeType: "AND",
		leftChild:  &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"},
		rightChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "Apache-2.0"}}
	parseAndCompare(t, "MIT AND Apache-2.0", nodeTree)

	// basic complex expression with plus
	nodeTree = &licenseNode{nodeType: "AND",
		leftChild:  &licenseNode{nodeType: "IDENTIFIER", identifier: "EPL-1.0", plus: true},
		rightChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "Apache-2.0"}}
	parseAndCompare(t, "EPL-1.0+ AND Apache-2.0", nodeTree)

	// basic complex expression with parens
	nodeTree = &licenseNode{nodeType: "AND",
		leftChild:  &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"},
		rightChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "Apache-2.0"}}
	parseAndCompare(t, "(MIT AND Apache-2.0)", nodeTree)

	// more complex expression with parens
	nodeTree = &licenseNode{nodeType: "OR",
		leftChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"},
		rightChild: &licenseNode{nodeType: "AND",
			leftChild:  &licenseNode{nodeType: "IDENTIFIER", identifier: "GPL-3.0-or-later"},
			rightChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "Apache-2.0"}}}
	parseAndCompare(t, "MIT OR (GPL-3.0-or-later AND Apache-2.0)", nodeTree)

	// multi-level complex expression with parens
	nodeTree = &licenseNode{nodeType: "AND",
		leftChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "MIT"},
		rightChild: &licenseNode{nodeType: "AND",
			leftChild: &licenseNode{nodeType: "OR",
				leftChild:  &licenseNode{nodeType: "IDENTIFIER", identifier: "Zlib"},
				rightChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "X11"}},
			rightChild: &licenseNode{nodeType: "IDENTIFIER", identifier: "Apache-2.0"}}}
	parseAndCompare(t, "MIT AND (Zlib OR X11) AND Apache-2.0", nodeTree)

}
