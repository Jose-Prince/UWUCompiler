package lib

type GrammarToken struct {
	Terminal    *string
	NonTerminal *string

	// Determines if this token is the '$' token at the end of a grammar.
	IsEnd bool
}

func NewEndToken() GrammarToken {
	return GrammarToken{
		IsEnd: true,
	}
}
func NewTerminalToken(val string) GrammarToken {
	return GrammarToken{
		Terminal: &val,
	}
}

func NewNonTerminalToken(val string) GrammarToken {
	return GrammarToken{
		NonTerminal: &val,
	}
}

type FirstFollowRow struct {
	First  Set[GrammarToken]
	Follow Set[GrammarToken]
}

type FirstFollowTable struct {
	table map[GrammarToken]FirstFollowRow
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
	return self.Terminal != nil
}

func (self *GrammarToken) IsNonTerminal() bool {
	return self.NonTerminal != nil
}

func (self *GrammarToken) Equal(other *GrammarToken) bool {
	if self.IsTerminal() && other.IsTerminal() {
		return *self.Terminal == *other.Terminal
	}

	if self.IsNonTerminal() && other.IsNonTerminal() {
		return *self.NonTerminal == *other.NonTerminal
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

func getFirsts(grammar *Grammar, table *FirstFollowTable) {
	alreadyEvaluatedFirsts := NewSet[GrammarToken]()
	table.AppendFirst(grammar.InitialSimbol, NewEndToken())

	evaluationStack := Stack[GrammarToken]{}
	for nonTerminal := range grammar.NonTerminals {
		evaluationStack.Push(nonTerminal)

		for evaluationStack.Peek().HasValue() {
			current := evaluationStack.Pop().GetValue()

			rulesWhereHead := getAllRulesWhereTokenIsHead(grammar, current)
			nonTerminalsFirsts := []GrammarToken{}

			for _, rule := range rulesWhereHead {
				firstFromProduction := rule.Production[0]

				if firstFromProduction.IsTerminal() {
					table.AppendFirst(nonTerminal, firstFromProduction)
				} else if !firstFromProduction.Equal(&nonTerminal) {
					nonTerminalsFirsts = append(nonTerminalsFirsts, firstFromProduction)
				}
			}

		}

	}
}

func getFollows(grammar *Grammar, table *FirstFollowTable) {
	for _, i := range grammar.Rules {
		head := i.Head
		production := i.Production

		if head.Equal(&grammar.InitialSimbol) {
			table(head).follow := add(table(head).follow, "$")
		}

		if len(production) == 1 {

		}
	}
}
