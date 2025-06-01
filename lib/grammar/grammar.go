package grammar

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Jose-Prince/UWUCompiler/lib"
	parsertypes "github.com/Jose-Prince/UWUCompiler/parserTypes"
)

type EpsilonString = lib.Optional[string]

type GrammarToken struct {
	Terminal    lib.Optional[EpsilonString]
	NonTerminal lib.Optional[string]

	// Determines if this token is the '$' token at the end of a grammar.
	IsEnd bool
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
	if self.IsEnd && other.IsEnd {
		return true
	}

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
	// Represents the final numeric id of the token.
	// You can use the file definition order,
	// so the first defined token will have id 0 and so on
	TokenIds map[GrammarToken]parsertypes.GrammarToken
}

func (g *Grammar) GetTokenByString(tokenStr string) GrammarToken {
	if tokenStr == "$" || tokenStr == "EOF" || tokenStr == "END" {
		return NewEndToken()
	}

	for _, terminal := range g.Terminals.ToSlice() {
		if terminal.String() == tokenStr {
			return terminal
		}
	}

	for _, nonTerminal := range g.NonTerminals.ToSlice() {
		if nonTerminal.String() == tokenStr {
			return nonTerminal
		}
	}

	if len(tokenStr) > 0 {
		firstChar := tokenStr[0]
		if firstChar >= 'A' && firstChar <= 'Z' {
			return NewNonTerminalToken(tokenStr)
		} else {
			return NewTerminalToken(tokenStr)
		}
	}

	return NewTerminalToken(tokenStr)
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

func (r1 GrammarRule) EqualRule(r2 *GrammarRule) bool {
	if !r1.Head.Equal(&r2.Head) {
		return false
	}

	if len(r1.Production) != len(r2.Production) {
		return false
	}

	for i := range r1.Production {
		if !r1.Production[i].Equal(&r2.Production[i]) {
			return false
		}
	}

	return true
}

func ParseYalFile(filename string) (Grammar, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Grammar{}, err
	}
	defer file.Close()

	var (
		terminals     = lib.NewSet[GrammarToken]()
		nonTerminals  = lib.NewSet[GrammarToken]()
		rules         []GrammarRule
		tokenIds      = make(map[GrammarToken]parsertypes.GrammarToken)
		initialSymbol GrammarToken
		foundStart    = false
	)

	terminals.Add(NewEndToken())

	scanner := bufio.NewScanner(file)
	mode := "header"
	tokenIdCounter := 0

	// Buffer to accumulate multi-line rules
	var currentRule strings.Builder
	var currentHead string
	inRule := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments (both // and /* */ style)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}

		// Handle section separator
		if line == "%%" {
			mode = "rules"
			continue
		}

		if mode == "header" {
			// Parse token declarations
			if strings.HasPrefix(line, "%token") {
				// Handle both single and multiple tokens on one line
				tokenLine := strings.TrimPrefix(line, "%token")
				tokenLine = strings.TrimSpace(tokenLine)
				parts := strings.Fields(tokenLine)

				for _, part := range parts {
					// Skip commented out tokens
					if strings.HasPrefix(part, "/*") || strings.HasSuffix(part, "*/") {
						continue
					}

					tok := NewTerminalToken(part)
					terminals.Add(tok)
					tokenIds[tok] = parsertypes.GrammarToken(tokenIdCounter)
					tokenIdCounter++
				}
			} else if strings.HasPrefix(line, "%start") {
				// Parse start symbol
				sym := strings.TrimSpace(strings.TrimPrefix(line, "%start"))
				initialSymbol = NewNonTerminalToken(sym)
				nonTerminals.Add(initialSymbol)
				foundStart = true
			}
		} else if mode == "rules" {
			// Handle rule parsing
			if strings.Contains(line, ":") {
				// Process any accumulated rule first
				if inRule && currentRule.Len() > 0 {
					processRule(currentHead, currentRule.String(), &rules, terminals, nonTerminals)
					currentRule.Reset()
				}

				// Start new rule
				parts := strings.SplitN(line, ":", 2)
				currentHead = strings.TrimSpace(parts[0])
				ruleBody := strings.TrimSpace(parts[1])

				// Add head to non-terminals
				headToken := NewNonTerminalToken(currentHead)
				nonTerminals.Add(headToken)

				// If there's no start symbol defined, use the first rule's head
				if !foundStart {
					initialSymbol = headToken
					foundStart = true
				}

				inRule = true
				if ruleBody != "" {
					currentRule.WriteString(ruleBody)
				}
			} else if inRule {
				// Continue accumulating rule body
				if line != "" {
					if currentRule.Len() > 0 {
						currentRule.WriteString(" ")
					}
					currentRule.WriteString(line)
				}
			}

			// Check if rule ends with semicolon
			if strings.HasSuffix(line, ";") {
				if inRule && currentRule.Len() > 0 {
					ruleStr := currentRule.String()
					if strings.HasSuffix(ruleStr, ";") {
						ruleStr = strings.TrimSuffix(ruleStr, ";")
					}
					processRule(currentHead, ruleStr, &rules, terminals, nonTerminals)
					currentRule.Reset()
					inRule = false
				}
			}
		}
	}

	// Process any remaining rule
	if inRule && currentRule.Len() > 0 {
		ruleStr := currentRule.String()
		if strings.HasSuffix(ruleStr, ";") {
			ruleStr = strings.TrimSuffix(ruleStr, ";")
		}
		processRule(currentHead, ruleStr, &rules, terminals, nonTerminals)
	}

	if err := scanner.Err(); err != nil {
		return Grammar{}, err
	}

	if !foundStart {
		return Grammar{}, fmt.Errorf("no start symbol defined and no rules found")
	}

	// Assign token IDs to non-terminals
	for nonTerminal := range nonTerminals {
		if _, exists := tokenIds[nonTerminal]; !exists {
			tokenIds[nonTerminal] = parsertypes.GrammarToken(tokenIdCounter)
			tokenIdCounter++
		}
	}

	// Add end token
	endToken := NewEndToken()
	tokenIds[endToken] = parsertypes.GrammarToken(tokenIdCounter)

	gram := Grammar{
		InitialSimbol: initialSymbol,
		Rules:         rules,
		Terminals:     terminals,
		NonTerminals:  nonTerminals,
		TokenIds:      tokenIds,
	}

	return gram, nil
}

// processRule processes a complete rule body and creates grammar rules
func processRule(headName, ruleBody string, rules *[]GrammarRule, terminals lib.Set[GrammarToken], nonTerminals lib.Set[GrammarToken]) {
	headToken := NewNonTerminalToken(headName)
	nonTerminals.Add(headToken)

	// Split by | for alternative productions
	alternatives := strings.Split(ruleBody, "|")

	for _, alt := range alternatives {
		alt = strings.TrimSpace(alt)
		if alt == "" {
			continue
		}

		production := []GrammarToken{}
		symbols := strings.Fields(alt)

		for _, sym := range symbols {
			sym = strings.TrimSpace(sym)
			if sym == "" {
				continue
			}

			var tok GrammarToken

			// Check for epsilon
			if sym == "ε" || sym == "epsilon" || sym == "EPSILON" {
				tok = CreateEpsilonToken()
			} else if isTerminal(sym, terminals) {
				tok = NewTerminalToken(sym)
			} else {
				// It's a non-terminal
				tok = NewNonTerminalToken(sym)
				nonTerminals.Add(tok)
			}

			production = append(production, tok)
		}

		// Handle empty production (epsilon)
		if len(production) == 0 {
			production = append(production, CreateEpsilonToken())
		}

		*rules = append(*rules, GrammarRule{
			Head:       headToken,
			Production: production,
		})
	}
}

func isTerminal(symbol string, terminals lib.Set[GrammarToken]) bool {
	testToken := NewTerminalToken(symbol)
	return terminals.Contains(testToken)
}

// Helper function to print grammar for debugging
func (g Grammar) PrintGrammar() {
	fmt.Println("=== GRAMMAR ===")
	fmt.Printf("Initial Symbol: %s\n", g.InitialSimbol.String())

	fmt.Println("\nTerminals:")
	for terminal := range g.Terminals {
		fmt.Printf("  %s\n", terminal.String())
	}

	fmt.Println("\nNon-Terminals:")
	for nonTerminal := range g.NonTerminals {
		fmt.Printf("  %s\n", nonTerminal.String())
	}

	fmt.Println("\nRules:")
	for i, rule := range g.Rules {
		fmt.Printf("  %d: %s\n", i, rule.ToString())
	}

	fmt.Println("\nToken IDs:")
	for token, id := range g.TokenIds {
		fmt.Printf("  %s -> %d\n", token.String(), id)
	}
}
