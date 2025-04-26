package lib

import "strings"

type GrammarToken struct {
	Terminal    Optional[string]
	NonTerminal Optional[string]

	// Determines if this token is the '$' token at the end of a grammar.
	IsEnd bool
}

func (self GrammarToken) String() string {
	b := strings.Builder{}
	b.WriteString("{ ")
	if self.IsTerminal() {
		b.WriteString("TERM: ")
		b.WriteString(self.Terminal.GetValue())
	} else if self.IsNonTerminal() {
		b.WriteString("NONT: ")
		b.WriteString(self.NonTerminal.GetValue())
	} else if self.IsEnd {
		b.WriteString("END: $")
	} else {
		b.WriteString("INVALID")
	}
	b.WriteString(" }")
	return b.String()
}

func NewEndToken() GrammarToken {
	return GrammarToken{
		IsEnd: true,
	}
}
func NewTerminalToken(val string) GrammarToken {
	return GrammarToken{
		NonTerminal: CreateNull[string](),
		Terminal:    CreateValue(val),
	}
}

func NewNonTerminalToken(val string) GrammarToken {
	return GrammarToken{
		NonTerminal: CreateValue(val),
		Terminal:    CreateNull[string](),
	}
}

func createEpsilonToken() GrammarToken {
	e := "ε"
	return NewTerminalToken(e)
}

func IsEpsilon(terminalToken GrammarToken) bool {
	return terminalToken.IsTerminal() && terminalToken.Terminal.GetValue() == "ε"
}

type FirstFollowRow struct {
	First  Set[GrammarToken]
	Follow Set[GrammarToken]
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
			First:  NewSet[GrammarToken](),
			Follow: NewSet[GrammarToken](),
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
			First:  NewSet[GrammarToken](),
			Follow: NewSet[GrammarToken](),
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

type Grammar struct {
	InitialSimbol GrammarToken
	Rules         []GrammarRule
	Terminals     Set[GrammarToken]
	NonTerminals  Set[GrammarToken]
}

func getFirstOfSequence(seq []GrammarToken, table *FirstFollowTable) Set[GrammarToken] {
	result := NewSet[GrammarToken]()

	for _, symbol := range seq {
		firstSet := table.table[symbol].First
		hasEpsilon := false

		for terminal := range firstSet {
			if IsEpsilon(terminal) {
				hasEpsilon = true
			} else {
				result.Add(terminal)
			}
		}

		if !hasEpsilon {
			return result
		}
	}

	result.Add(createEpsilonToken())
	return result
}

func GetFirsts(grammar *Grammar, table *FirstFollowTable) {
	alreadyEvaluatedFirsts := NewSet[GrammarToken]()

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

func getFirstFor(grammar *Grammar, table *FirstFollowTable, current *GrammarToken, alreadyEvaluated *Set[GrammarToken]) {
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

	table.AppendFollow(grammar.InitialSimbol, NewEndToken())

	for changed {
		changed = false

		for _, rule := range grammar.Rules {
			head := rule.Head
			production := rule.Production

			for i := 0; i < len(production); i++ {
				B := production[i]

				if !B.IsNonTerminal() {
					continue
				}

				if i+1 < len(production) {
					beta := production[i+1:]
					firstOfBeta := getFirstOfSequence(beta, table)

					for terminal := range firstOfBeta {
						if !terminal.IsEnd && !(terminal.IsNonTerminal() && IsEpsilon(terminal)) {

							if table.table[B].Follow.Add_(terminal) {
								changed = true
							}
						}
					}

					if firstOfBeta.Contains(createEpsilonToken()) {
						for terminal := range table.table[head].Follow {
							if table.table[B].Follow.Add_(terminal) {
								changed = true
							}
						}
					}
				} else {
					for terminal := range table.table[head].Follow {
						if table.table[B].Follow.Add_(terminal) {
							changed = true
						}
					}
				}
			}
		}
	}
}
