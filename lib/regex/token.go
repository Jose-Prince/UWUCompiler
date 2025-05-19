package regex

import (
	"fmt"
	"strings"

	"github.com/Jose-Prince/UWULexer/lib"
)

// Represents a Regular Expression operator
type Operator int

const (
	OR           Operator = iota // |
	AND                          // Concatenation operator
	ZERO_OR_MANY                 // *
	ONE_OR_MANY                  // +
	OPTIONAL                     // ?
	LEFT_PAREN                   // (
	RIGHT_PAREN                  // )
)

func (self *Operator) String() string {
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
	}

	return displayOp
}

// Serves to append extra metadata to a Regex pattern.
type DummyInfo struct {
	// Original string regex associated with this dummy.
	Regex string
	// The code to execute once the Regex pattern is identified.
	Code string
	// Used to break ties when parsing tokens!
	//
	// The lower the number the higher the priority!
	Priority uint
}

func (self *DummyInfo) Equals(other *DummyInfo) bool {
	return self.Regex == other.Regex
}

// Represents a rune that may be an epsilon
type EpsilonRune = lib.Optional[rune]

// Represents a token.
// It can either be a value, an operator or a dummy token.
// If value is null then it should have an operator value, otherwise a value should be provided!
type RX_Token struct {
	// If the token is an operator this will be not nil.
	operator lib.Optional[Operator]
	// If value is nil then this token is not a value.
	// If the optional doesn't have a value then the value is epsilon.
	// If the optional has a value then this token has the value of the rune.
	value lib.Optional[EpsilonRune]
	// If the token is a dummy token this will be not nil.
	dummy lib.Optional[DummyInfo]
}

func (self *RX_Token) GetValue() EpsilonRune {
	if !self.IsValue() {
		panic(fmt.Sprintf("The token `%s` is not a value!", self.String()))
	}
	return self.value.GetValue()
}

func (self *RX_Token) GetOperator() Operator {
	if !self.IsOperator() {
		panic(fmt.Sprintf("The token `%s` is not an operator!", self.String()))
	}
	return self.operator.GetValue()
}

func (self *RX_Token) GetDummy() DummyInfo {
	if !self.IsDummy() {
		panic(fmt.Sprintf("The token `%s` is not a dummy token!", self.String()))
	}

	return self.dummy.GetValue()
}

func (self *RX_Token) IsValue() bool {
	return self.value.HasValue()
}

func (self *RX_Token) IsOperator() bool {
	return self.operator.HasValue()
}

func (self *RX_Token) IsDummy() bool {
	return self.dummy.HasValue()
}

func CreateOperatorToken(t Operator) RX_Token {
	return RX_Token{
		operator: lib.CreateValue(t),
	}
}

func CreateValueToken(r rune) RX_Token {
	return RX_Token{
		value: lib.CreateValue(lib.CreateValue(r)),
	}
}

func CreateEpsilonToken() RX_Token {
	return RX_Token{
		value: lib.CreateValue(lib.CreateNull[rune]()),
	}
}

func CreateDummyToken(info DummyInfo) RX_Token {
	return RX_Token{
		dummy: lib.CreateValue(info),
	}
}

func (self *RX_Token) Equals(other *RX_Token) bool {
	if self.IsOperator() && other.IsOperator() {
		selfOp := self.GetOperator()
		otherOp := other.GetOperator()
		return selfOp == otherOp

	} else if self.IsValue() && other.IsValue() {
		val := self.GetValue()
		otherVal := other.GetValue()

		if val.HasValue() && otherVal.HasValue() {
			return val.GetValue() == otherVal.GetValue()
		} else {
			return val.HasValue() == otherVal.HasValue()
		}

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
	return !self.dummy.HasValue() && !self.operator.HasValue() && !self.value.HasValue()
}

func TokenStreamToString(stream []RX_Token) string {
	b := strings.Builder{}
	for _, elem := range stream {
		b.WriteString(elem.String())
		b.WriteRune('\n')
	}

	return b.String()
}

func (self *RX_Token) String() string {
	if self.IsOperator() {
		op := self.GetOperator()
		return fmt.Sprintf("{ opr = %s }", op.String())
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
		return fmt.Sprintf("{ dummy = `%s` }", self.GetDummy().Regex)
	}

	return "{ undefined token type }"
}
