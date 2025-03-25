package lib

import (
	"fmt"
	"math"
	"os"
)

// func generatePositions(afd *AFD) map[string][2]float64 {
//     pi_value := 0.0
//     centerX, centerY := 200.0, 300.0
//     radius := 120.0
//     positions := make(map[string][2]float64)

//     // Recopilar todos los estados (no solo los que tienen transiciones)
//     allStates := make(map[string]bool)

//     // Agregar estados desde las transiciones
//     for from, transitions := range afd.Transitions {
//         allStates[from] = true
//         for _, to := range transitions {
//             allStates[to] = true
//         }
//     }

//     // Agregar estados que no aparezcan en las transiciones (como el estado final)
//     for state := range afd.AcceptanceStates {
//         allStates[state] = true
//     }

//     // Convertir a slice para iterar de manera ordenada
//     stateList := make([]string, 0, len(allStates))
//     for state := range allStates {
//         stateList = append(stateList, state)
//     }

//     // Generar posiciones circulares para todos los estados
//     for _, state := range stateList {
//         sin, cos := math.Sincos(pi_value)
//         positions[state] = [2]float64{centerX + radius*cos, centerY + radius*sin}
//         pi_value += 2 * math.Pi / float64(len(stateList))
//     }

//     return positions
// }

// func (afd *AFD) ToSVG() string {
//     statePositions := generatePositions(afd)

//     svg := `<svg width="50%" height="50%" viewBox="0 0 600 600" xmlns="http://www.w3.org/2000/svg">`
//     svg += `<rect width="100%" height="100%" fill="white"/>`

//     radius := 30.0 // Radio de los nodos

//     // Mapa para agrupar etiquetas por transición
//     transitionLabels := make(map[[2]string][]string)

//     // Recopilar transiciones
//     for from, transitions := range afd.Transitions {
//         for input, to := range transitions {
//             key := [2]string{from, to}
//             transitionLabels[key] = append(transitionLabels[key], input)
//         }
//     }

//     // Dibujar los estados
//     for state, pos := range statePositions {
//         x, y := pos[0], pos[1]
//         fill := "white"
//         if afd.AcceptanceStates.Contains(state) {
//             fill = "lightgreen"
//         }
//         svg += fmt.Sprintf(`<circle cx="%f" cy="%f" r="%f" stroke="black" stroke-width="2" fill="%s"/>`, x, y, radius, fill)
//         svg += fmt.Sprintf(`<text x="%f" y="%f" font-size="16" text-anchor="middle" fill="black">%s</text>`, x, y+5, state)
//     }

//     // Dibujar flecha de estado inicial
//     initialState := afd.InitialState
//     if pos, exists := statePositions[initialState]; exists {
//         x, y := pos[0], pos[1]
//         arrowStartX := x - radius - 20 // Punto de inicio de la flecha
//         arrowEndX := x - radius        // Punto donde toca el círculo

//         svg += fmt.Sprintf(
//             `<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="black" stroke-width="2" marker-end="url(#arrow)"/>`,
//             arrowStartX, y, arrowEndX, y)
//     }

//     // Dibujar las transiciones sin sobrescribir etiquetas
//     for key, inputs := range transitionLabels {
//         from, to := key[0], key[1]
//         x1, y1 := statePositions[from][0], statePositions[from][1]
//         x2, y2 := statePositions[to][0], statePositions[to][1]

//         labels := strings.Join(inputs, ", ") // Combinar etiquetas

//         if from == to {
//             // 🔄 Dibujar loop más abajo
//             loopRadius := 40.0

//             svg += fmt.Sprintf(
//                 `<path d="M %f %f C %f %f, %f %f, %f %f" stroke="black" stroke-width="2" fill="none" marker-end="url(#arrow)"/>`,
//                 x1, y1-radius,
//                 x1+loopRadius+10, y1-radius-loopRadius-10,
//                 x1-loopRadius-10, y1-radius-loopRadius-10,
//                 x1, y1-radius)

//             // Etiqueta en el loop
//             svg += fmt.Sprintf(
//                 `<text x="%f" y="%f" font-size="14" fill="black">%s</text>`,
//                 x1, y1-radius-loopRadius-15, labels)
//         } else {
//             // 🔀 Dibujar transiciones normales
//             dx := float64(x2 - x1)
//             dy := float64(y2 - y1)
//             dist := math.Sqrt(dx*dx + dy*dy)

//             if dist > 0 {
//                 unitDx := dx / dist
//                 unitDy := dy / dist
//                 x1 += unitDx * float64(radius)
//                 y1 += unitDy * float64(radius)
//                 x2 -= unitDx * float64(radius)
//                 y2 -= unitDy * float64(radius)
//             }

//             // Línea de transición
//             svg += fmt.Sprintf(
//                 `<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="black" stroke-width="2" marker-end="url(#arrow)"/>`,
//                 x1, y1, x2, y2)

//             // Etiqueta de transición combinada
//             svg += fmt.Sprintf(
//                 `<text x="%f" y="%f" font-size="14" fill="black">%s</text>`,
//                 (x1+x2)/2, (y1+y2)/2 - 5, labels)
//         }
//     }

//     // Definir flechas más grandes
//     svg += `<defs>
//         <marker id="arrow" markerWidth="15" markerHeight="15" refX="12" refY="6" orient="auto">
//             <path d="M 0 0 L 12 6 L 0 12 z" fill="black"/>
//         </marker>
//     </defs>`

//     svg += `</svg>`
//     return svg
// }

// func GenerateHTML(svgContent, outputHTML string) error {
//     htmlContent := fmt.Sprintf(`
// 	<!DOCTYPE html>
// 	<html lang="en">
// 	<head>
// 		<meta charset="UTF-8">
// 		<meta name="viewport" content="width=device-width, initial-scale=1.0">
// 		<title>Visualización AFD</title>
// 	</head>
// 	<body>
// 		<h2>Autómata Finito Determinista</h2>
// 		%s
// 	</body>
// 	</html>`, svgContent)

//		return os.WriteFile(outputHTML, []byte(htmlContent), 0644)
//	}

func treeHeight(bst *BST, index int) int {
	if index == -1 {
		return 0
	}
	leftHeight := treeHeight(bst, bst.nodes[index].left)
	rightHeight := treeHeight(bst, bst.nodes[index].right)
	return 1 + int(math.Max(float64(leftHeight), float64(rightHeight)))
}

// Mapea cada nodo a su nivel en el árbol
func assignLevels(bst *BST, index int, depth int, levelMap map[int][]int) {
	if index == -1 {
		return
	}
	levelMap[depth] = append(levelMap[depth], index)
	assignLevels(bst, bst.nodes[index].left, depth+1, levelMap)
	assignLevels(bst, bst.nodes[index].right, depth+1, levelMap)
}

// Calcula posiciones centradas para los nodos
func calculatePositions(bst *BST, width, height float64) map[int][2]float64 {
	if len(bst.nodes) == 0 {
		return nil
	}

	positions := make(map[int][2]float64)
	treeDepth := treeHeight(bst, len(bst.nodes)-1)
	levelMap := make(map[int][]int)
	assignLevels(bst, len(bst.nodes)-1, 0, levelMap)

	levelSpacing := height / float64(treeDepth+1)
	for depth, nodes := range levelMap {
		numNodes := len(nodes)
		baseX := width / float64(numNodes+1)

		for i, index := range nodes {
			x := baseX * float64(i+1)
			y := float64(depth+1) * levelSpacing
			positions[index] = [2]float64{x, y}
		}
	}
	return positions
}

// Genera el SVG con los nodos centrados
func GenerateBSTSVG(bst *BST) string {
	width := 800.0
	height := 400.0
	positions := calculatePositions(bst, width, height)

	svg := fmt.Sprintf(`<svg width="%f" height="%f" xmlns="http://www.w3.org/2000/svg">`, width, height)
	svg += `<rect width="100%" height="100%" fill="white"/>`

	// Dibujar conexiones entre nodos
	for i, node := range bst.nodes {
		if node.left != -1 {
			x1, y1 := positions[i][0], positions[i][1]
			x2, y2 := positions[node.left][0], positions[node.left][1]
			svg += fmt.Sprintf(`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="black"/>`, x1, y1, x2, y2)
		}
		if node.right != -1 {
			x1, y1 := positions[i][0], positions[i][1]
			x2, y2 := positions[node.right][0], positions[node.right][1]
			svg += fmt.Sprintf(`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="black"/>`, x1, y1, x2, y2)
		}
	}

	// Dibujar nodos
	for i, node := range bst.nodes {

		value := ""
		if node.Val.IsValue() {
			opt := node.Val.GetValue()
			if opt.HasValue() {
				value = string(opt.GetValue())
			}
		} else if node.Val.IsOperator() {
			opt := node.Val.GetOperator()
			value = opt.String()
		} else {
			value = node.Val.GetDummy().Regex
		}

		x, y := positions[i][0], positions[i][1]
		svg += fmt.Sprintf(`<circle cx="%f" cy="%f" r="20" stroke="black" fill="white"/>`, x, y)
		svg += fmt.Sprintf(`<text x="%f" y="%f" font-size="14" text-anchor="middle" fill="black">%s</text>`, x, y+5, value)
	}

	svg += `</svg>`
	return svg
}

// Genera un archivo HTML con el SVG
func GenerateHTMLBST(svgContent, filename string) error {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="es">
	<head>
		<meta charset="UTF-8">
		<title>Visualización BST</title>
	</head>
	<body>
		<h2>Árbol Binario de Búsqueda</h2>
		%s
	</body>
	</html>`, svgContent)

	return os.WriteFile(filename, []byte(html), 0644)
}
