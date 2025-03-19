package lib

import (
	"fmt"
	"strings"
)

// Represents a Regular Expression operator
type Operator int

const (
	OR            Operator = iota // |
	AND                           // Concatenation operator
	ZERO_OR_MANY                  // *
	ONE_OR_MANY                   // +
	OPTIONAL                      // ?
	LEFT_PAREN                    // (
	RIGHT_PAREN                   // )
	LEFT_BRACKET                  // [
	RIGHT_BRACKET                 // ]
	SET_NEGATION                  // [^
)

// Serves to append extra metadata to a Regex pattern.
type DummyInfo struct {
	// The code to execute once the Regex pattern is identified.
	code string
}

func (self *DummyInfo) Equals(other *DummyInfo) bool {
	return self.code == other.code
}

// Represents a token.
// It can either be a value, an operator or a dummy token.
// If value is null then it should have an operator value, otherwise a value should be provided!
type RX_Token struct {
	// If the token is an operator this will be not nil.
	operator *Operator
	// If the value is nil then this token is not a value.
	// If the optional doesn't have a value then the value is epsilon.
	// If the optional has a value then this token has the value of the rune.
	value *Optional[rune]
	// If the token is a dummy token this will be not nil.
	dummy *DummyInfo
}

func (self *RX_Token) GetValue() Optional[rune] {
	if !self.IsValue() {
		panic(fmt.Sprintf("The token `%s` is not a value!", self.ToString()))
	}
	return *self.value
}

func (self *RX_Token) GetOperator() Operator {
	if !self.IsOperator() {
		panic(fmt.Sprintf("The token `%s` is not an operator!", self.ToString()))
	}
	return *self.operator
}

func (self *RX_Token) GetDummy() DummyInfo {
	if !self.IsDummy() {
		panic(fmt.Sprintf("The token `%s` is not a dummy token!", self.ToString()))
	}

	return *self.dummy
}

func (self *RX_Token) IsValue() bool {
	return self.value != nil
}

func (self *RX_Token) IsOperator() bool {
	return self.operator != nil
}

func (self *RX_Token) IsDummy() bool {
	return self.dummy != nil
}

func CreateOperatorToken(t Operator) RX_Token {
	return RX_Token{
		operator: &t,
	}
}

func CreateValueToken(value rune) RX_Token {
	val := CreateValue(value)
	return RX_Token{
		value: &val,
	}
}

func CreateEpsilonToken() RX_Token {
	val := CreateNull[rune]()
	return RX_Token{
		value: &val,
	}
}

func (self *RX_Token) Equals(other *RX_Token) bool {
	if self.IsOperator() && other.IsOperator() {
		return *self.operator == *other.operator

	} else if self.IsValue() && other.IsValue() {
		val := self.GetValue()
		otherVal := other.GetValue()
		return (&val).Equals(&otherVal)

	} else if self.IsDummy() && other.IsDummy() {
		dum := self.GetDummy()
		otherDum := other.GetDummy()
		return (&dum).Equals(&otherDum)

	} else {
		selfUninitialized := self.IsUninitialized()
		otherUninitialized := other.IsUninitialized()

		// An uninitialized token compared to another uninitialized token should be equal!
		// Kinda the same logic that NULL == NULL
		return selfUninitialized && otherUninitialized
	}
}

func (self *RX_Token) IsUninitialized() bool {
	return self.dummy == nil && self.operator == nil && self.value == nil
}

func TokenStreamToString(stream []RX_Token) string {
	b := strings.Builder{}
	for _, elem := range stream {
		b.WriteString(elem.ToString())
		b.WriteByte(' ')
	}

	return b.String()
}

func (self *Operator) ToString() string {
	displayOp := "invalid"

	switch *self {
	case OR:
		displayOp = "|"
	case AND:
		displayOp = "."
	case ZERO_OR_MANY:
		displayOp = "*"
	case ONE_OR_MANY:
		displayOp = "+"
	case OPTIONAL:
		displayOp = "?"
	case LEFT_PAREN:
		displayOp = "("
	case RIGHT_PAREN:
		displayOp = ")"
	case LEFT_BRACKET:
		displayOp = "["
	case RIGHT_BRACKET:
		displayOp = "]"
	case SET_NEGATION:
		displayOp = "[^"
	}

	return displayOp
}

func (self *RX_Token) ToString() string {
	if self.IsOperator() {
		op := self.GetOperator()
		return fmt.Sprintf("{ opr = %s }", op.ToString())
	}

	if self.IsValue() {
		val := "epsilon"
		opt := self.GetValue()
		if opt.HasValue() {
			val = string(opt.GetValue())
		}

		return fmt.Sprintf("{ val = %s }", val)
	}

	if self.IsDummy() {
		return fmt.Sprintf("{ dummy = `%s` }", self.GetDummy().code)
	}

	return "{ undefined token type }"
}
