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

func getNodeTreeString(nodeTree *licenseNode) string {
	if nodeTree == nil {
		return "_"
	}

	var label string
	if nodeTree.nodeType == "IDENTIFIER" {
		label = "IDENTIFIER:" + nodeTree.identifier
	} else {
		label = nodeTree.nodeType
	}

	return "[ " + label + " " +
		getNodeTreeString(nodeTree.leftChild) + " " +
		getNodeTreeString(nodeTree.rightChild) + " ]"
}

const (
	nextLeft int = iota
	nextRight
	nextNone
)

func getParseError(exprTokens []exprToken, parentNode *licenseNode, i int, tok exprToken) error {
	return fmt.Errorf("invalid parse tree: encountered %s(%s) (token %d) for %v; full tree: %v", tok.t, tok.identifier, i, exprTokens, parentNode)
}

func nodeTreeIsFull(parentNode *licenseNode) bool {
	if parentNode.nodeType == "IDENTIFIER" {
		return true
	}
	if parentNode.nodeType == "AND" || parentNode.nodeType == "OR" {
		return parentNode.leftChild != nil && nodeTreeIsFull(parentNode.leftChild) && parentNode.rightChild != nil && nodeTreeIsFull(parentNode.rightChild)
	}
	return false
}

func parseTokens(exprTokens []exprToken) (*licenseNode, error) {
	curParentNode := &licenseNode{}
	curNode := curParentNode
	next := nextNone

	for i, tok := range exprTokens {
		switch tok.t {
		case "IDENTIFIER":
			if curNode.nodeType == "" {
				// this node is free, so fill it in
				curNode.nodeType = "IDENTIFIER"
				curNode.identifier = tok.identifier
			} else if next == nextLeft && curNode.leftChild == nil && (curNode.nodeType == "AND" || curNode.nodeType == "OR" || curNode.nodeType == "LPAREN") {
				// left child is free and can be the next identifier
				curNode.leftChild = &licenseNode{nodeType: "IDENTIFIER", identifier: tok.identifier, parent: curNode}
				curNode = curNode.leftChild
				next = nextNone
			} else if next == nextRight && curNode.rightChild == nil && (curNode.nodeType == "AND" || curNode.nodeType == "OR") {
				// right child is free and can be the next identifier
				curNode.rightChild = &licenseNode{nodeType: "IDENTIFIER", identifier: tok.identifier, parent: curNode}
				curNode = curNode.rightChild
				next = nextNone
			} else {
				// if we got here, it isn't a valid parse tree
				return curParentNode, getParseError(exprTokens, curParentNode, i, tok)
			}

		case "AND":
			if nodeTreeIsFull(curNode) && (curNode.nodeType == "IDENTIFIER" || curNode.nodeType == "AND" || curNode.nodeType == "OR") {
				// insert new node and push existing one down
				newNode := &licenseNode{nodeType: "AND", parent: curNode.parent, leftChild: curNode}
				if curNode.parent == nil {
					curParentNode = newNode
				} else {
					if curNode.parent.leftChild == curNode {
						curNode.parent.leftChild = newNode
					} else {
						curNode.parent.rightChild = newNode
					}
					curNode.parent = newNode
				}
				curNode = newNode
				next = nextRight
			} else {
				// if we got here, it isn't a valid parse tree
				return curParentNode, getParseError(exprTokens, curParentNode, i, tok)
			}

		case "OR":
			if nodeTreeIsFull(curNode) && (curNode.nodeType == "IDENTIFIER" || curNode.nodeType == "AND" || curNode.nodeType == "OR") {
				// insert new node and push existing one down
				newNode := &licenseNode{nodeType: "OR", parent: curNode.parent, leftChild: curNode}
				if curNode.parent == nil {
					curParentNode = newNode
				} else {
					if curNode.parent.leftChild == curNode {
						curNode.parent.leftChild = newNode
					} else {
						curNode.parent.rightChild = newNode
					}
				}
				curNode.parent = newNode
				curNode = newNode
				next = nextRight
			} else {
				// if we got here, it isn't a valid parse tree
				return curParentNode, getParseError(exprTokens, curParentNode, i, tok)
			}

		case "PLUS":
			if curNode.nodeType == "IDENTIFIER" && curNode.plus == false && next == nextNone {
				curNode.plus = true
			} else {
				// if we got here, it isn't a valid parse tree
				return curParentNode, getParseError(exprTokens, curParentNode, i, tok)
			}

		case "LPAREN":
			if curNode.nodeType == "" {
				// at the beginning
				curNode.nodeType = "LPAREN"
				next = nextLeft
			} else if curNode.nodeType == "LPAREN" && curNode.leftChild == nil && next == nextLeft {
				// at a sub-paren on left; create another LPAREN child
				curNode.leftChild = &licenseNode{nodeType: "LPAREN", parent: curNode}
				curNode = curNode.leftChild
				next = nextLeft
			} else if (curNode.nodeType == "AND" || curNode.nodeType == "OR") && curNode.rightChild == nil && next == nextRight {
				// at a paren following an open AND or OR; create another LPAREN child
				// FIXME is there a situation here where we would need to go to its leftChild instead?
				curNode.rightChild = &licenseNode{nodeType: "LPAREN", parent: curNode}
				curNode = curNode.rightChild
				next = nextLeft
			} else {
				// if we got here, it isn't a valid parse tree
				return curParentNode, getParseError(exprTokens, curParentNode, i, tok)
			}

		case "RPAREN":
			// find lowermost LPAREN node
			found := false
			for iter := curNode; iter != nil; iter = iter.parent {
				// fmt.Printf("===> IN RPAREN, iter is %v\n", iter)
				if iter.nodeType == "LPAREN" {
					// roll up and replace LPAREN node with its child
					child := iter.leftChild
					child.parent = iter.parent
					if iter.parent == nil {
						// at root, so change parent pointer
						curParentNode = child
					} else {
						// not at root, so re-route
						if iter.parent.rightChild == iter {
							iter.parent.rightChild = child
						} else {
							iter.parent.leftChild = child
						}
					}
					curNode = child
					next = nextNone
					found = true
					break
				}
			}
			if !found {
				// if we got here, it isn't a valid parse tree
				return curParentNode, getParseError(exprTokens, curParentNode, i, tok)
			}
		}

		// fmt.Printf("=> parsing token %d - %v...\n", i, tok)
		// fmt.Printf("Tree: %v\n", getNodeTreeString(curParentNode))
		// fmt.Printf("curNode: %v\n", curNode)
		// fmt.Printf("curNode.parent: %v\n", curNode.parent)
		// fmt.Printf("\n")
	}

	// FIXME check whether tree is full
	return curParentNode, nil
}
