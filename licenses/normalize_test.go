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
