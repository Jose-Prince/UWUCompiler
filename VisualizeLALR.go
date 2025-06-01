package main

import (
	"fmt"
	"math"
	"os"

	"github.com/Jose-Prince/UWUCompiler/lib/grammar"
)

func GenerateHTML(a grammar.Automata, filename string) error {
	const (
		nodeRadius = 30
		centerX    = 800
		centerY    = 500
		radius     = 350 // Radio del círculo
	)

	// Abrimos el archivo
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Autómata LALR</title></head>
<body>
<h1>Autómata LALR</h1>
<svg width="1600" height="1000" style="border:1px solid #ccc; font-family: monospace;">`)

	fmt.Fprintln(f, `<defs>
	<marker id="arrow" markerWidth="10" markerHeight="10" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
		<path d="M0,0 L0,6 L9,3 z" fill="black" />
	</marker>
</defs>`)

	nodePositions := make(map[int][2]float64)
	n := len(a.Nodes)
	i := 0
	for id := range a.Nodes {
		angle := 2 * math.Pi * float64(i) / float64(n)
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)
		nodePositions[id] = [2]float64{x, y}
		i++
	}

	// Dibujamos transiciones
	for fromID, state := range a.Nodes {
		from := nodePositions[fromID]
		for symbol, toID := range state.Productions {
			to := nodePositions[toID]

			dx := to[0] - from[0]
			dy := to[1] - from[1]
			angle := math.Atan2(dy, dx)

			// Rectángulo es de 180x(rectHeight). Usamos mitades para calcular offset
			offsetX := 90 * math.Cos(angle)
			offsetY := 40 * math.Sin(angle)

			startX := from[0] + offsetX
			startY := from[1] + offsetY
			endX := to[0] - offsetX
			endY := to[1] - offsetY

			// Línea con flecha
			fmt.Fprintf(f, `<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="black" marker-end="url(#arrow)" />`,
				startX, startY, endX, endY)

			// Texto de transición
			labelX := (startX + endX) / 2
			labelY := (startY+endY)/2 - 5
			fmt.Fprintf(f, `<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12">%s</text>`,
				labelX, labelY, symbol)
		}
	}

	// Dibujamos nodos y sus producciones
	for id, pos := range nodePositions {
		color := "#add8e6"
		if a.Nodes[id].Initial {
			color = "#ffcccb"
		}
		if a.Nodes[id].Accept {
			color = "#90ee90"
		}

		// Calcular altura dinámica basada en la cantidad de items
		lineHeight := 16.0
		padding := 10.0
		numLines := float64(len(a.Nodes[id].Items) + 1) // +1 para el índice
		rectHeight := padding*2 + lineHeight*numLines
		rectWidth := 180.0

		x := pos[0] - rectWidth/2
		y := pos[1] - rectHeight/2

		// Rectángulo
		fmt.Fprintf(f, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" stroke="black" rx="10" ry="10" />`,
			x, y, rectWidth, rectHeight, color)

		// Texto interno
		textX := pos[0]
		textY := y + padding + lineHeight

		// Índice del nodo
		fmt.Fprintf(f, `<text x="%.2f" y="%.2f" text-anchor="middle" font-size="13" font-weight="bold">I%d</text>`,
			textX, textY, id)

		// Reglas
		textY += lineHeight
		for _, item := range a.Nodes[id].Items {
			ruleStr := formatItem(item)
			fmt.Fprintf(f, `<text x="%.2f" y="%.2f" text-anchor="middle" font-size="11">%s</text>`,
				textX, textY, ruleStr)
			textY += lineHeight
		}
	}

	fmt.Fprintln(f, `</svg></body></html>`)
	return nil
}

func formatItem(item grammar.AutomataItem) string {
	var out string

	// Encabezado de la regla
	if item.Rule.Head.NonTerminal.HasValue() {
		out += item.Rule.Head.NonTerminal.GetValue()
	} else {
		out += item.Rule.Head.Terminal.GetValue().GetValue()
	}
	out += " →"

	for i, tok := range item.Rule.Production {
		if i == item.DotPosition {
			out += " •"
		}
		out += " " + tokenToString(tok)
	}
	if item.DotPosition == len(item.Rule.Production) {
		out += " •"
	}

	// Lookahead
	out += " , {"
	for i, look := range item.Lookahead {
		if i > 0 {
			out += ", "
		}
		out += tokenToString(look)
	}
	out += "}"

	return out
}

func tokenToString(tok grammar.GrammarToken) string {
	if tok.IsEnd {
		return "$"
	}
	if tok.NonTerminal.HasValue() {
		return tok.NonTerminal.GetValue()
	}
	if tok.Terminal.HasValue() {
		return tok.Terminal.GetValue().GetValue()
	}
	return "ε"
}
