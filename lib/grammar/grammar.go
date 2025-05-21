package grammar

import (
	"fmt"
	"strings"

	"github.com/Jose-Prince/UWULexer/lib"
)

type EpsilonString = lib.Optional[string]

type GrammarToken struct {
	Terminal    lib.Optional[EpsilonString]
	NonTerminal lib.Optional[string]

	// Determines if this token is the '$' token at the end of a grammar.
	IsEnd bool
	// Determines if this token is ',' for the look ahead.
	IsLookAhead bool
	// Determines if this token is '.' for making closure in the grammar
	IsPointer bool
}

func (self GrammarToken) String() string {
	b := strings.Builder{}
	b.WriteString("{")
	if self.IsTerminal() {
		b.WriteString("TERM: ")
		val := self.Terminal.GetValue()
		if val.HasValue() {
			b.WriteString(val.GetValue())
		} else {
			b.WriteRune('ε')
		}
	} else if self.IsNonTerminal() {
		b.WriteString("NONT: ")
		b.WriteString(self.NonTerminal.GetValue())
	} else if self.IsEnd {
		b.WriteString("END: $")
	} else {
		b.WriteString("INVALID")
	}
	b.WriteString("}")
	return b.String()
}

func NewEndToken() GrammarToken {
	return GrammarToken{
		IsEnd: true,
	}
}

func NewLookAheadToken() GrammarToken {
	return GrammarToken{
		IsLookAhead: true,
	}
}

func NewPointerToken() GrammarToken {
	return GrammarToken{
		IsPointer: true,
	}
}

func NewTerminalToken(val string) GrammarToken {
	return GrammarToken{
		NonTerminal: lib.CreateNull[string](),
		Terminal:    lib.CreateValue(lib.CreateValue(val)),
	}
}

func NewNonTerminalToken(val string) GrammarToken {
	return GrammarToken{
		NonTerminal: lib.CreateValue(val),
		Terminal:    lib.CreateNull[EpsilonString](),
	}
}

func CreateEpsilonToken() GrammarToken {
	return GrammarToken{
		NonTerminal: lib.CreateNull[string](),
		Terminal:    lib.CreateValue(lib.CreateNull[string]()),
	}
}

func IsEpsilon(terminalToken GrammarToken) bool {
	return terminalToken.IsTerminal() && !terminalToken.Terminal.GetValue().HasValue()
}

type FirstFollowRow struct {
	First  lib.Set[GrammarToken]
	Follow lib.Set[GrammarToken]
}

func (self FirstFollowRow) String() string {
	b := strings.Builder{}
	b.WriteString("{ ")
	b.WriteString("\tFirsts: ")
	b.WriteString(self.First.String())
	b.WriteString("\tFollows: ")
	b.WriteString(self.Follow.String())
	b.WriteString(" }")
	return b.String()
}

type FirstFollowTable struct {
	table map[GrammarToken]FirstFollowRow
}

func NewFirstFollowTable() FirstFollowTable {
	table := make(map[GrammarToken]FirstFollowRow)
	return FirstFollowTable{
		table: table,
	}
}

func (self *FirstFollowTable) AppendFirst(key GrammarToken, val GrammarToken) {
	if _, found := self.table[key]; !found {
		self.table[key] = FirstFollowRow{
			First:  lib.NewSet[GrammarToken](),
			Follow: lib.NewSet[GrammarToken](),
		}
	}

	row := self.table[key]
	first := row.First
	first.Add(val)

	row.First = first
	self.table[key] = row
}

func (self *FirstFollowTable) AppendFollow(key GrammarToken, val GrammarToken) {
	if _, found := self.table[key]; !found {
		self.table[key] = FirstFollowRow{
			First:  lib.NewSet[GrammarToken](),
			Follow: lib.NewSet[GrammarToken](),
		}
	}

	row := self.table[key]
	follow := row.Follow
	follow.Add(val)

	row.Follow = follow
	self.table[key] = row
}

func (self *GrammarToken) IsTerminal() bool {
	return self.Terminal.HasValue()
}

func (self *GrammarToken) IsNonTerminal() bool {
	return self.NonTerminal.HasValue()
}

func (self *GrammarToken) Equal(other *GrammarToken) bool {
	if self.IsTerminal() && other.IsTerminal() {
		return self.Terminal.Equals(&other.Terminal)
	}

	if self.IsNonTerminal() && other.IsNonTerminal() {
		return self.NonTerminal.Equals(&other.NonTerminal)
	}

	return false
}

type GrammarRule struct {
	Head       GrammarToken
	Production []GrammarToken
}

func (self GrammarRule) ToString() string {
	prodStr := ""
	for _, p := range self.Production {
		prodStr += p.String() + " "
	}
	return fmt.Sprintf("%s -> %s", self.Head.String(), prodStr)
}

type Grammar struct {
	InitialSimbol GrammarToken
	Rules         []GrammarRule
	Terminals     lib.Set[GrammarToken]
	NonTerminals  lib.Set[GrammarToken]
}

func GetFirsts(grammar *Grammar, table *FirstFollowTable) {
	alreadyEvaluatedFirsts := lib.NewSet[GrammarToken]()

	for nonTerminal := range grammar.NonTerminals {
		getFirstFor(grammar, table, &nonTerminal, &alreadyEvaluatedFirsts)
	}
}

func getAllRulesWhereTokenIsHead(grammar *Grammar, token *GrammarToken) []GrammarRule {
	rules := []GrammarRule{}
	for _, rule := range grammar.Rules {
		if rule.Head.Equal(token) {
			rules = append(rules, rule)
		}
	}
	return rules
}

func getFirstFor(grammar *Grammar, table *FirstFollowTable, current *GrammarToken, alreadyEvaluated *lib.Set[GrammarToken]) {
	if !alreadyEvaluated.Add(*current) {
		return
	}

	rulesWhereHead := getAllRulesWhereTokenIsHead(grammar, current)
	for _, rule := range rulesWhereHead {
		firstFromProduction := rule.Production[0]

		if firstFromProduction.IsTerminal() {
			table.AppendFirst(*current, firstFromProduction)

		} else if !firstFromProduction.Equal(current) {
			getFirstFor(grammar, table, &firstFromProduction, alreadyEvaluated)

			for evl := range table.table[firstFromProduction].First {
				table.AppendFirst(*current, evl)
			}
		}
	}
}

func GetFollows(grammar *Grammar, table *FirstFollowTable) {
	changed := true

	table.AppendFollow(grammar.InitialSimbol, NewTerminalToken("$"))

	for val := range grammar.NonTerminals {
		if _, exists := table.table[val]; !exists {
			table.table[val] = FirstFollowRow{
				First:  lib.NewSet[GrammarToken](),
				Follow: lib.NewSet[GrammarToken](),
			}
		}
	}

	for changed {
		changed = false

		for _, rule := range grammar.Rules {
			A := rule.Head
			production := rule.Production

			for i, B := range production {
				follow := table.table[B].Follow

				if !B.IsNonTerminal() {
					continue
				}

				if i+1 < len(production) {
					beta := production[i+1:]

					for _, terminal := range beta {
						if terminal.IsTerminal() {

							if follow.Add(terminal) {
								changed = true
							}
						} else {
							newTerminal := derivateNonTerminal(terminal, grammar)

							if follow.Add(newTerminal) {
								changed = true
							}
						}
					}

					break
				} else {
					for terminal := range table.table[A].Follow {
						if follow.Add(terminal) {
							changed = true
						}
					}
				}
			}
		}
	}
}

func derivateNonTerminal(token GrammarToken, grammar *Grammar) GrammarToken {

	if token.IsTerminal() {
		return token
	}

	for _, rule := range grammar.Rules {
		recursive := false

		// Check if rule isnt recursive
		for _, producction := range rule.Production {
			if producction.Equal(&token) {
				recursive = true
			}
		}

		if rule.Head.Equal(&token) && len(rule.Production) == 1 && !recursive {
			return derivateNonTerminal(rule.Production[0], grammar)
		}
	}

	return CreateEpsilonToken()
}

func (g *Grammar) First(token GrammarToken) []GrammarToken {
	result := make(map[string]GrammarToken)

	if token.IsTerminal() {
		result[token.String()] = token
		return mapToSlice(result)
	}

	for _, rule := range g.Rules {
		if rule.Head.Equal(&token) {
			for i := 0; i < len(rule.Production); i++ {
				symbol := rule.Production[i]
				if symbol.IsTerminal() {
					result[symbol.String()] = symbol
					break
				}

				firsts := g.First(symbol)
				hasEpsilon := false
				for _, f := range firsts {
					if f.String() == "ε" {
						hasEpsilon = true
					} else {
						result[f.String()] = f
					}
				}

				if !hasEpsilon {
					break
				}

				if i == len(rule.Production)-1 && hasEpsilon {
					result["ε"] = NewEndToken()
				}
			}
		}
	}

	return mapToSlice(result)
}

func mapToSlice(m map[string]GrammarToken) []GrammarToken {
	slice := make([]GrammarToken, 0, len(m))
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}
