// Copyright The Linux Foundation
// Portions of this file that are original to The Linux Foundation are
// licensed under the Apache License, version 2.0.

// Some portions of this file are derived from the lexmachine "sensors"
// example. lexmachine is under the BSD-3-Clause license with the following
// copyright statement for the full library:
// Copyright (c) 2014-2017 All rights reserved; portions may be owned by:
//     * Tim Henderson
//     * Case Western Reserve University
//     * Google Inc.

// SPDX-License-Identifier: Apache-2.0 AND BSD-3-Clause

package licenses

import (
	"fmt"

	"github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

type exprToken struct {
	t          string
	identifier string
}

var tokens = []string{
	"AND",
	"OR",
	"PLUS",
	"WITH",
	"LPAREN",
	"RPAREN",
	"IDENTIFIER",
}

var tokenMap map[string]int
var lexer *lexmachine.Lexer

func init() {
	tokenMap = make(map[string]int)
	for tokenID, tokenName := range tokens {
		tokenMap[tokenName] = tokenID
	}

	lexer = createLexer()
}

func createLexer() *lexmachine.Lexer {
	getToken := func(tokenType int) lexmachine.Action {
		return func(s *lexmachine.Scanner, m *machines.Match) (interface{}, error) {
			return s.Token(tokenType, string(m.Bytes), m), nil
		}
	}
	skip := func(scan *lexmachine.Scanner, match *machines.Match) (interface{}, error) {
		return nil, nil
	}
	lexer := lexmachine.NewLexer()
	lexer.Add([]byte("AND"), getToken(tokenMap["AND"]))
	lexer.Add([]byte("OR"), getToken(tokenMap["OR"]))
	lexer.Add([]byte(`\+`), getToken(tokenMap["PLUS"]))
	lexer.Add([]byte("WITH"), getToken(tokenMap["WITH"]))
	lexer.Add([]byte(`\(`), getToken(tokenMap["LPAREN"]))
	lexer.Add([]byte(`\)`), getToken(tokenMap["RPAREN"]))
	lexer.Add([]byte(`[a-zA-Z0-9\-\.]+`), getToken(tokenMap["IDENTIFIER"]))
	lexer.Add([]byte(`#[^\n]*`), skip)
	lexer.Add([]byte(`( |\t|\f)+`), skip)
	lexer.Add([]byte(`\\\n`), skip)
	lexer.Add([]byte(`\n|\r|\n\r`), skip)

	err := lexer.Compile()
	if err != nil {
		panic(fmt.Errorf("couldn't create lexer; %v", err))
	}

	return lexer
}

func getTokens(expression string) ([]exprToken, error) {
	var exprTokens []exprToken

	scanner, err := lexer.Scanner([]byte(expression))
	if err != nil {
		return nil, fmt.Errorf("error lexing for tokens: %v", err)
	}

	for tk, err, eof := scanner.Next(); !eof; tk, err, eof = scanner.Next() {
		if err != nil {
			return nil, fmt.Errorf("error lexing for tokens mid-string: %v", err)
		}
		token := tk.(*lexmachine.Token)
		eToken := exprToken{t: tokens[token.Type], identifier: token.Value.(string)}
		exprTokens = append(exprTokens, eToken)
	}

	return exprTokens, nil
}

type licenseNode struct {
	nodeType   string
	identifier string
	plus       bool
	parent     *licenseNode
	leftChild  *licenseNode
	rightChild *licenseNode
}

func getExpressionFromNodeTree(nodeTreeParent *licenseNode) (string, error) {
	output, err := getExpressionFromNodeTreeHelper(nodeTreeParent)
	if err != nil {
		return "", err
	}

	// check for, and remove if present, outermost parens
	if output[:1] == "(" {
		return output[1 : len(output)-1], nil
	}

	return output, nil
}

func getExpressionFromNodeTreeHelper(nodeTreeParent *licenseNode) (string, error) {
	switch nodeTreeParent.nodeType {
	case "IDENTIFIER":
		if nodeTreeParent.plus {
			return nodeTreeParent.identifier + "+", nil
		}
		return nodeTreeParent.identifier, nil
	case "AND":
		leftResult, err := getExpressionFromNodeTreeHelper(nodeTreeParent.leftChild)
		if err != nil {
			return "", fmt.Errorf("error getting left child from AND: %v", err)
		}
		rightResult, err := getExpressionFromNodeTreeHelper(nodeTreeParent.rightChild)
		if err != nil {
			return "", fmt.Errorf("error getting right child from AND: %v", err)
		}

		return "(" + leftResult + " AND " + rightResult + ")", nil
	case "OR":
		leftResult, err := getExpressionFromNodeTreeHelper(nodeTreeParent.leftChild)
		if err != nil {
			return "", fmt.Errorf("error getting left child from OR: %v", err)
		}
		rightResult, err := getExpressionFromNodeTreeHelper(nodeTreeParent.rightChild)
		if err != nil {
			return "", fmt.Errorf("error getting right child from OR: %v", err)
		}

		return "(" + leftResult + " OR " + rightResult + ")", nil
	case "WITH":
		leftResult, err := getExpressionFromNodeTreeHelper(nodeTreeParent.leftChild)
		if err != nil {
			return "", fmt.Errorf("error getting left child from WITH: %v", err)
		}
		rightResult, err := getExpressionFromNodeTreeHelper(nodeTreeParent.rightChild)
		if err != nil {
			return "", fmt.Errorf("error getting right child from WITH: %v", err)
		}

		return "(" + leftResult + " WITH " + rightResult + ")", nil
	}

	return "", fmt.Errorf("error getting expression with invalid nodeType: %v", nodeTreeParent.nodeType)
}

// func parseTokens(exprTokens []exprToken) (*licenseNode, error) {

// }
