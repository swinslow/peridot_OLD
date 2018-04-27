// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package licenses

import "fmt"

type parsedLicenseNodeType int

const (
	// NodeIdentifier indicates a node with a single license identifier
	NodeIdentifier parsedLicenseNodeType = iota

	// NodeAnd indicates a node that is the "AND" of two subexpressions
	NodeAnd

	// NodeOr indicates a node that is the "OR" of two subexpressions
	NodeOr

	// NodeWith indicates a node that contains a license (possibly with PLUS)
	// as a left child, and an exception as a right child
	NodeWith

	// NodePlus indicates a node that has a NodeIdentifier node as its left
	// child (nil as right), interpreted as "this version or later"
	NodePlus
)

// ParsedLicenseNode is the externally-visible representation for the contents
// of a parsed SPDX license expression node tree.
type ParsedLicenseNode struct {
	NodeType   parsedLicenseNodeType
	Identifier string
	LeftChild  *ParsedLicenseNode
	RightChild *ParsedLicenseNode
}

func convertNodeTreeToPLN(nodeTree *licenseNode) (*ParsedLicenseNode, error) {
	if nodeTree == nil {
		return nil, nil
	}

	switch nodeTree.nodeType {
	case "IDENTIFIER":
		if nodeTree.plus == true {
			// create a Plus node with an ID node as its left child
			child := &ParsedLicenseNode{NodeType: NodeIdentifier, Identifier: nodeTree.identifier}
			return &ParsedLicenseNode{NodeType: NodePlus, LeftChild: child}, nil
		}
		return &ParsedLicenseNode{NodeType: NodeIdentifier, Identifier: nodeTree.identifier}, nil

	case "AND":
		leftChild, err := convertNodeTreeToPLN(nodeTree.leftChild)
		if err != nil {
			return nil, err
		}
		rightChild, err := convertNodeTreeToPLN(nodeTree.rightChild)
		if err != nil {
			return nil, err
		}
		return &ParsedLicenseNode{NodeType: NodeAnd, LeftChild: leftChild, RightChild: rightChild}, nil

	case "OR":
		leftChild, err := convertNodeTreeToPLN(nodeTree.leftChild)
		if err != nil {
			return nil, err
		}
		rightChild, err := convertNodeTreeToPLN(nodeTree.rightChild)
		if err != nil {
			return nil, err
		}
		return &ParsedLicenseNode{NodeType: NodeOr, LeftChild: leftChild, RightChild: rightChild}, nil

	case "WITH":
		leftChild, err := convertNodeTreeToPLN(nodeTree.leftChild)
		if err != nil {
			return nil, err
		}
		rightChild, err := convertNodeTreeToPLN(nodeTree.rightChild)
		if err != nil {
			return nil, err
		}
		return &ParsedLicenseNode{NodeType: NodeWith, LeftChild: leftChild, RightChild: rightChild}, nil

	}

	return nil, fmt.Errorf("got invalid nodeType %v converting internal nodeTree to ParsedLicenseNode", nodeTree.nodeType)
}

// GetNodesForExpression lexes and parses an SPDX license expression, and
// returns a pointer to a ParsedLicenseNode tree representing that
// expression.
func GetNodesForExpression(expr string) (*ParsedLicenseNode, error) {
	// break expression into tokens
	tokens, err := getTokens(expr)
	if err != nil {
		return nil, fmt.Errorf("error lexing expression: %v", err)
	}

	// parse tokens
	nodeTree, err := parseTokens(tokens)
	if err != nil {
		return nil, fmt.Errorf("error parsing expression: %v", err)
	}

	// convert parsed tokens into PLN
	parsedNodes, err := convertNodeTreeToPLN(nodeTree)
	if err != nil {
		return nil, fmt.Errorf("error converting to node format: %v", err)
	}

	return parsedNodes, nil
}
