// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package licenses

import "fmt"

// ParsedLicenseNodeType represents the various types of license nodes
// that can be obtained from parsing an SPDX license expression.
type ParsedLicenseNodeType int

const (
	// NodeIdentifier indicates a node with a single license identifier
	NodeIdentifier ParsedLicenseNodeType = iota

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

	// NodeError indicates an invalid node type or other error
	NodeError
)

// ParsedLicenseNode is the externally-visible representation for the contents
// of a parsed SPDX license expression node tree.
type ParsedLicenseNode struct {
	NodeType   ParsedLicenseNodeType
	Expression string
	LeftChild  *ParsedLicenseNode
	RightChild *ParsedLicenseNode
}

func isConjunction(nt ParsedLicenseNodeType) bool {
	return nt == NodeAnd || nt == NodeOr || nt == NodeWith
}

func getChildExpression(pln *ParsedLicenseNode) string {
	if isConjunction(pln.NodeType) {
		return "(" + pln.Expression + ")"
	}

	return pln.Expression
}

func convertNodeTreeToPLN(nodeTree *licenseNode) (*ParsedLicenseNode, error) {
	var leftExpr, rightExpr string

	if nodeTree == nil {
		return nil, nil
	}

	switch nodeTree.nodeType {
	case "IDENTIFIER":
		if nodeTree.plus == true {
			// create a Plus node with an ID node as its left child
			child := &ParsedLicenseNode{
				NodeType:   NodeIdentifier,
				Expression: nodeTree.identifier}

			return &ParsedLicenseNode{
				NodeType:   NodePlus,
				Expression: nodeTree.identifier + "+",
				LeftChild:  child}, nil
		}
		return &ParsedLicenseNode{
			NodeType:   NodeIdentifier,
			Expression: nodeTree.identifier}, nil

	case "AND":
		leftChild, err := convertNodeTreeToPLN(nodeTree.leftChild)
		if err != nil {
			return nil, err
		}
		leftExpr = getChildExpression(leftChild)

		rightChild, err := convertNodeTreeToPLN(nodeTree.rightChild)
		if err != nil {
			return nil, err
		}
		rightExpr = getChildExpression(rightChild)

		return &ParsedLicenseNode{
			NodeType:   NodeAnd,
			Expression: leftExpr + " AND " + rightExpr,
			LeftChild:  leftChild,
			RightChild: rightChild}, nil

	case "OR":
		leftChild, err := convertNodeTreeToPLN(nodeTree.leftChild)
		if err != nil {
			return nil, err
		}
		leftExpr = getChildExpression(leftChild)

		rightChild, err := convertNodeTreeToPLN(nodeTree.rightChild)
		if err != nil {
			return nil, err
		}
		rightExpr = getChildExpression(rightChild)

		return &ParsedLicenseNode{
			NodeType:   NodeOr,
			Expression: leftExpr + " OR " + rightExpr,
			LeftChild:  leftChild,
			RightChild: rightChild}, nil

	case "WITH":
		leftChild, err := convertNodeTreeToPLN(nodeTree.leftChild)
		if err != nil {
			return nil, err
		}
		leftExpr = getChildExpression(leftChild)

		rightChild, err := convertNodeTreeToPLN(nodeTree.rightChild)
		if err != nil {
			return nil, err
		}
		rightExpr = getChildExpression(rightChild)

		return &ParsedLicenseNode{
			NodeType:   NodeWith,
			Expression: leftExpr + " WITH " + rightExpr,
			LeftChild:  leftChild,
			RightChild: rightChild}, nil

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
