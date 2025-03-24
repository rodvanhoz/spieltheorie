package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Choice representa a escolha de uma estratégia (Cooperar ou Trair)
type Choice int

const (
	Cooperate Choice = iota
	Defect
)

// Strategy define uma interface para as estratégias
type Strategy interface {
	NextMove(round int, opponentMoves []Choice) Choice
	Name() string
}

// TitForTat: Coopera na primeira rodada, depois imita o último movimento do oponente
type TitForTat struct{}

func (s TitForTat) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		return Cooperate
	}
	return opponentMoves[len(opponentMoves)-1]
}
func (s TitForTat) Name() string { return "Tit-for-Tat" }

// Random: Escolhe aleatoriamente entre cooperar e trair
type Random struct{}

func (s Random) NextMove(round int, opponentMoves []Choice) Choice {
	if rand.Intn(2) == 0 {
		return Cooperate
	}
	return Defect
}
func (s Random) Name() string { return "Random" }

// TidemanChieruzzi: Variação de Tit-for-Tat com perdão baseado no histórico
type TidemanChieruzzi struct{}

func (s TidemanChieruzzi) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		return Cooperate
	}
	// Se o oponente traiu na última rodada, verifica o histórico
	lastMove := opponentMoves[len(opponentMoves)-1]
	if lastMove == Defect {
		// Conta o número de traições e cooperações recentes (últimas 5 rodadas)
		recentDefects := 0
		recentMoves := opponentMoves[max(0, len(opponentMoves)-5):]
		for _, move := range recentMoves {
			if move == Defect {
				recentDefects++
			}
		}
		// Perdoa se o oponente traiu menos de 50% das vezes recentemente
		if recentDefects < len(recentMoves)/2 {
			return Cooperate
		}
	}
	return lastMove
}
func (s TidemanChieruzzi) Name() string { return "Tideman & Chieruzzi" }

// Nydegger: Usa uma sequência inicial para testar o oponente
type Nydegger struct{}

func (s Nydegger) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 {
		return Cooperate
	}
	if round == 1 {
		return Defect
	}
	if round == 2 {
		return Cooperate
	}
	// Após as 3 primeiras rodadas, decide com base nas respostas do oponente
	if round == 3 {
		// Se o oponente cooperou nas 3 primeiras rodadas, coopera
		if opponentMoves[0] == Cooperate && opponentMoves[1] == Cooperate && opponentMoves[2] == Cooperate {
			return Cooperate
		}
		return Defect
	}
	// Depois disso, age como Tit-for-Tat
	return opponentMoves[len(opponentMoves)-1]
}
func (s Nydegger) Name() string { return "Nydegger" }

// Grofman: Coopera na maioria das vezes, trai a cada 5 rodadas
type Grofman struct{}

func (s Grofman) NextMove(round int, opponentMoves []Choice) Choice {
	if round%5 == 0 { // Trai a cada 5 rodadas
		return Defect
	}
	return Cooperate
}
func (s Grofman) Name() string { return "Grofman" }

// Shubik: Tit-for-Tat com punição prolongada (2 rodadas de traição)
type Shubik struct {
	defectCount int
}

func (s *Shubik) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		s.defectCount = 0
		return Cooperate
	}
	if s.defectCount > 0 {
		s.defectCount--
		return Defect
	}
	lastMove := opponentMoves[len(opponentMoves)-1]
	if lastMove == Defect {
		s.defectCount = 1 // Pune por 2 rodadas (1 adicional, já que esta rodada é uma traição)
		return Defect
	}
	return Cooperate
}
func (s Shubik) Name() string { return "Shubik" }

// SteinRapoport: Tit-for-Tat com perdão aleatório
type SteinRapoport struct{}

func (s SteinRapoport) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		return Cooperate
	}
	lastMove := opponentMoves[len(opponentMoves)-1]
	if lastMove == Defect {
		// 20% de chance de perdoar uma traição
		if rand.Float64() < 0.2 {
			return Cooperate
		}
	}
	return lastMove
}
func (s SteinRapoport) Name() string { return "Stein & Rapoport" }

// Friedman: Grim Trigger (trai para sempre após a primeira traição)
type Friedman struct {
	triggered bool
}

func (s *Friedman) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		s.triggered = false
		return Cooperate
	}
	if s.triggered {
		return Defect
	}
	lastMove := opponentMoves[len(opponentMoves)-1]
	if lastMove == Defect {
		s.triggered = true
		return Defect
	}
	return Cooperate
}
func (s Friedman) Name() string { return "Friedman" }

// Davis: Coopera por 10 rodadas, depois age como Tit-for-Tat
type Davis struct{}

func (s Davis) NextMove(round int, opponentMoves []Choice) Choice {
	if round < 10 {
		return Cooperate
	}
	return opponentMoves[len(opponentMoves)-1]
}
func (s Davis) Name() string { return "Davis" }

// Graaskamp: Analisa a proporção de traições do oponente
type Graaskamp struct{}

func (s Graaskamp) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		return Cooperate
	}
	// Calcula a proporção de traições do oponente
	defectCount := 0
	for _, move := range opponentMoves {
		if move == Defect {
			defectCount++
		}
	}
	proportion := float64(defectCount) / float64(len(opponentMoves))
	// Se o oponente traiu mais de 50% das vezes, trai; caso contrário, coopera
	if proportion > 0.5 {
		return Defect
	}
	return Cooperate
}
func (s Graaskamp) Name() string { return "Graaskamp" }

// Downing: Estima se o oponente responde melhor a cooperação ou traição
type Downing struct {
	coopScore, defectScore int
}

func (s *Downing) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		s.coopScore = 0
		s.defectScore = 0
		return Cooperate
	}
	// Atualiza pontuações com base nas respostas do oponente
	lastMove := opponentMoves[len(opponentMoves)-1]
	if lastMove == Cooperate {
		s.coopScore += 1
	} else {
		s.defectScore += 1
	}
	// Escolhe a ação que maximiza a resposta de cooperação do oponente
	if s.coopScore > s.defectScore {
		return Cooperate
	}
	return Defect
}
func (s Downing) Name() string { return "Downing" }

// Feld: Aumenta a probabilidade de trair ao longo do jogo
type Feld struct{}

func (s Feld) NextMove(round int, opponentMoves []Choice) Choice {
	// Probabilidade de trair aumenta linearmente com o número de rodadas
	probDefect := float64(round) / 200.0 // Ajuste para 200 rodadas como referência
	if probDefect > 1.0 {
		probDefect = 1.0
	}
	if rand.Float64() < probDefect {
		return Defect
	}
	return Cooperate
}
func (s Feld) Name() string { return "Feld" }

// Joss: Tit-for-Tat com 10% de chance de trair
type Joss struct{}

func (s Joss) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		return Cooperate
	}
	// 10% de chance de trair, independentemente do oponente
	if rand.Float64() < 0.1 {
		return Defect
	}
	return opponentMoves[len(opponentMoves)-1]
}
func (s Joss) Name() string { return "Joss" }

// Tullock: Coopera na maioria das vezes, trai ocasionalmente
type Tullock struct{}

func (s Tullock) NextMove(round int, opponentMoves []Choice) Choice {
	// 5% de chance de trair para testar o oponente
	if rand.Float64() < 0.05 {
		return Defect
	}
	return Cooperate
}
func (s Tullock) Name() string { return "Tullock" }

// NameWithheld: Variação de Tit-for-Tat com 5% de chance de trair
type NameWithheld struct{}

func (s NameWithheld) NextMove(round int, opponentMoves []Choice) Choice {
	if round == 0 || len(opponentMoves) == 0 {
		return Cooperate
	}
	// 5% de chance de trair
	if rand.Float64() < 0.05 {
		return Defect
	}
	return opponentMoves[len(opponentMoves)-1]
}
func (s NameWithheld) Name() string { return "Name Withheld" }

// Game representa o estado do jogo
type Game struct {
	strategyA, strategyB Strategy
	rounds               int
	scores               [2]int
	movesA, movesB       []Choice
}

// NewGame cria um novo jogo
func NewGame(strategyA, strategyB Strategy, rounds int) *Game {
	return &Game{
		strategyA: strategyA,
		strategyB: strategyB,
		rounds:    rounds,
		scores:    [2]int{0, 0},
		movesA:    make([]Choice, 0, rounds),
		movesB:    make([]Choice, 0, rounds),
	}
}

// PlayRound joga uma rodada e atualiza os pontos
func (g *Game) PlayRound(round int) {
	moveA := g.strategyA.NextMove(round, g.movesB)
	moveB := g.strategyB.NextMove(round, g.movesA)

	g.movesA = append(g.movesA, moveA)
	g.movesB = append(g.movesB, moveB)

	// Calcula pontuação
	if moveA == Cooperate && moveB == Cooperate {
		g.scores[0] += 7
		g.scores[1] += 7
	} else if moveA == Cooperate && moveB == Defect {
		g.scores[0] += 0
		g.scores[1] += 10
	} else if moveA == Defect && moveB == Cooperate {
		g.scores[0] += 10
		g.scores[1] += 0
	} else { // Ambos traem
		g.scores[0] += 1
		g.scores[1] += 1
	}
}

// moveToSymbol converte a escolha em um símbolo visual
func moveToSymbol(move Choice) string {
	if move == Cooperate {
		return "✅"
	}
	return "❌"
}

// max é uma função auxiliar para evitar índices negativos
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Result representa o resultado de uma estratégia no modo "todos contra todos"
type Result struct {
	name  string
	score int
}

// runAllAgainstAll executa o modo "todos contra todos" e retorna os resultados
func runAllAgainstAll(strategies []Strategy, rounds int) []Result {
	// Mapa para armazenar a pontuação total de cada estratégia
	totalScores := make(map[string]int)

	// Cada estratégia enfrenta todas as outras (incluindo a si mesma)
	for _, stratA := range strategies {
		for _, stratB := range strategies {
			// Cria instâncias frescas das estratégias para evitar estado compartilhado
			var strategyA, strategyB Strategy
			switch stratA.Name() {
			case "Shubik":
				strategyA = &Shubik{}
			case "Friedman":
				strategyA = &Friedman{}
			case "Downing":
				strategyA = &Downing{}
			default:
				strategyA = stratA
			}
			switch stratB.Name() {
			case "Shubik":
				strategyB = &Shubik{}
			case "Friedman":
				strategyB = &Friedman{}
			case "Downing":
				strategyB = &Downing{}
			default:
				strategyB = stratB
			}

			// Executa o jogo entre strategyA e strategyB
			game := NewGame(strategyA, strategyB, rounds)
			for round := 0; round < rounds; round++ {
				game.PlayRound(round)
			}

			// Adiciona os pontos ao total de cada estratégia
			totalScores[stratA.Name()] += game.scores[0]
			totalScores[stratB.Name()] += game.scores[1]
		}
	}

	// Converte os resultados para uma lista de Result
	results := make([]Result, 0, len(strategies))
	for name, score := range totalScores {
		results = append(results, Result{name: name, score: score})
	}

	// Ordena os resultados por pontuação (maior para menor)
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	return results
}

func main() {
	// Seed para escolhas aleatórias
	rand.Seed(time.Now().UnixNano())

	// Cria a aplicação Fyne
	myApp := app.New()
	myWindow := myApp.NewWindow("Spieltheorie - Teoria dos Jogos")
	myWindow.Resize(fyne.NewSize(800, 600))

	// Lista de estratégias disponíveis
	strategies := []Strategy{
		TitForTat{},
		Random{},
		TidemanChieruzzi{},
		Nydegger{},
		Grofman{},
		&Shubik{},
		SteinRapoport{},
		&Friedman{},
		Davis{},
		Graaskamp{},
		&Downing{},
		Feld{},
		Joss{},
		Tullock{},
		NameWithheld{},
	}

	// Lista de nomes das estratégias para os dropdowns
	strategyNames := make([]string, len(strategies))
	for i, s := range strategies {
		strategyNames[i] = s.Name()
	}

	// Tela inicial: escolha entre modo normal e "todos contra todos"
	welcomeLabel := widget.NewLabel("Bem-vindo ao Spieltheorie!")
	welcomeLabel.Alignment = fyne.TextAlignCenter

	normalModeButton := widget.NewButton("Modo Normal", func() {
		// Tela do modo normal
		strategyASelect := widget.NewSelect(strategyNames, func(value string) {})
		strategyASelect.SetSelected(strategyNames[0])
		strategyBSelect := widget.NewSelect(strategyNames, func(value string) {})
		strategyBSelect.SetSelected(strategyNames[1])

		roundsEntry := widget.NewEntry()
		roundsEntry.SetPlaceHolder("Digite o número de rodadas")

		// Barra de progresso para o progresso das rodadas
		progressBar := widget.NewProgressBar()
		progressBar.Min = 0
		progressBar.Max = 1

		// Tabela para exibir o histórico das rodadas
		type roundData struct {
			round  int
			moveA  string
			moveB  string
			scoreA int
			scoreB int
		}
		roundsHistory := make([]roundData, 0)

		// Cria a tabela
		table := widget.NewTable(
			func() (int, int) {
				return len(roundsHistory), 5 // 5 colunas: Rodada, Move A, Move B, Score A, Score B
			},
			func() fyne.CanvasObject {
				return widget.NewLabel("")
			},
			func(cell widget.TableCellID, o fyne.CanvasObject) {
				label := o.(*widget.Label)
				data := roundsHistory[cell.Row]
				switch cell.Col {
				case 0:
					label.SetText(fmt.Sprintf("%d", data.round))
				case 1:
					label.SetText(data.moveA)
				case 2:
					label.SetText(data.moveB)
				case 3:
					label.SetText(fmt.Sprintf("%d", data.scoreA))
				case 4:
					label.SetText(fmt.Sprintf("%d", data.scoreB))
				}
			},
		)
		// Define os cabeçalhos da tabela
		table.CreateHeader = func() fyne.CanvasObject {
			return widget.NewLabel("")
		}
		table.UpdateHeader = func(cell widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			switch cell.Col {
			case 0:
				label.SetText("Rodada")
			case 1:
				label.SetText("Jogada A")
			case 2:
				label.SetText("Jogada B")
			case 3:
				label.SetText("Pontuação A")
			case 4:
				label.SetText("Pontuação B")
			}
		}
		// Define larguras das colunas
		table.SetColumnWidth(0, 80)
		table.SetColumnWidth(1, 100)
		table.SetColumnWidth(2, 100)
		table.SetColumnWidth(3, 100)
		table.SetColumnWidth(4, 100)

		// Define um tamanho mínimo para a tabela (ex.: 10 linhas visíveis)
		table.MinSize()
		tableContainer := container.NewVScroll(table)
		tableContainer.SetMinSize(fyne.NewSize(500, 300)) // Ajusta para mostrar ~10 linhas

		// Label para o resultado final
		resultLabel := widget.NewLabel("")
		resultLabel.Wrapping = fyne.TextWrapWord

		startButton := widget.NewButton("Iniciar Jogo", func() {
			rounds, err := strconv.Atoi(roundsEntry.Text)
			if err != nil || rounds <= 0 {
				resultLabel.SetText("Por favor, insira um número de rodadas válido!")
				return
			}

			// Encontra as estratégias selecionadas
			var strategyA, strategyB Strategy
			for _, s := range strategies {
				if s.Name() == strategyASelect.Selected {
					switch s.Name() {
					case "Shubik":
						strategyA = &Shubik{}
					case "Friedman":
						strategyA = &Friedman{}
					case "Downing":
						strategyA = &Downing{}
					default:
						strategyA = s
					}
				}
				if s.Name() == strategyBSelect.Selected {
					switch s.Name() {
					case "Shubik":
						strategyB = &Shubik{}
					case "Friedman":
						strategyB = &Friedman{}
					case "Downing":
						strategyB = &Downing{}
					default:
						strategyB = s
					}
				}
			}

			// Atualiza os cabeçalhos da tabela com os nomes das estratégias
			table.UpdateHeader(widget.TableCellID{Row: -1, Col: 1}, widget.NewLabel(strategyA.Name()))
			table.UpdateHeader(widget.TableCellID{Row: -1, Col: 2}, widget.NewLabel(strategyB.Name()))

			// Limpa o histórico
			roundsHistory = roundsHistory[:0]
			table.Refresh()

			// Configura a barra de progresso
			progressBar.Max = float64(rounds)
			progressBar.Value = 0
			progressBar.Refresh()

			// Executa o jogo
			game := NewGame(strategyA, strategyB, rounds)
			for i := 0; i < rounds; i++ {
				game.PlayRound(i)

				// Adiciona a rodada ao histórico
				roundsHistory = append(roundsHistory, roundData{
					round:  i + 1,
					moveA:  moveToSymbol(game.movesA[i]),
					moveB:  moveToSymbol(game.movesB[i]),
					scoreA: game.scores[0],
					scoreB: game.scores[1],
				})

				// Atualiza a tabela e a barra de progresso
				progressBar.SetValue(float64(i + 1))
				table.Refresh()

				// Rola para a última linha
				if len(roundsHistory) > 0 {
					table.ScrollTo(widget.TableCellID{Row: len(roundsHistory) - 1, Col: 0})
				}

				time.Sleep(100 * time.Millisecond) // Pausa para visualização
			}

			// Resultado final
			var output strings.Builder
			output.WriteString("Resultado Final:\n")
			output.WriteString(fmt.Sprintf("%s: %d pontos\n", strategyA.Name(), game.scores[0]))
			output.WriteString(fmt.Sprintf("%s: %d pontos\n", strategyB.Name(), game.scores[1]))
			if game.scores[0] > game.scores[1] {
				output.WriteString(fmt.Sprintf("Vencedor: %s!\n", strategyA.Name()))
			} else if game.scores[1] > game.scores[0] {
				output.WriteString(fmt.Sprintf("Vencedor: %s!\n", strategyB.Name()))
			} else {
				output.WriteString("Empate!\n")
			}

			resultLabel.SetText(output.String())
		})

		// Layout do modo normal
		content := container.NewVBox(
			widget.NewLabel("Escolha a Estratégia A:"),
			strategyASelect,
			widget.NewLabel("Escolha a Estratégia B:"),
			strategyBSelect,
			widget.NewLabel("Número de Rodadas:"),
			roundsEntry,
			startButton,
			widget.NewLabel("Progresso:"),
			progressBar,
			widget.NewSeparator(),
			widget.NewLabel("Histórico das Rodadas:"),
			tableContainer,
			widget.NewSeparator(),
			resultLabel,
		)

		scroll := container.NewVScroll(content)
		myWindow.SetContent(scroll)
	})

	allModeButton := widget.NewButton("Modo Todos Contra Todos", func() {
		// Tela do modo "todos contra todos"
		roundsEntry := widget.NewEntry()
		roundsEntry.SetPlaceHolder("Digite o número de rodadas")

		outputLabel := widget.NewLabel("Resultado aparecerá aqui...")
		outputLabel.Wrapping = fyne.TextWrapWord

		startButton := widget.NewButton("Iniciar Torneio", func() {
			rounds, err := strconv.Atoi(roundsEntry.Text)
			if err != nil || rounds <= 0 {
				outputLabel.SetText("Por favor, insira um número de rodadas válido!")
				return
			}

			outputLabel.SetText("Processando...")
			fyne.CurrentApp().Driver().CanvasForObject(outputLabel).Refresh(outputLabel)

			// Executa o torneio
			results := runAllAgainstAll(strategies, rounds)

			// Exibe os resultados
			var output strings.Builder
			output.WriteString("Resultados Finais (ordenados por pontuação):\n")
			output.WriteString("------------------------------------------\n")
			for i, result := range results {
				output.WriteString(fmt.Sprintf("%d. %s: %d pontos\n", i+1, result.name, result.score))
			}

			outputLabel.SetText(output.String())
		})

		// Layout do modo "todos contra todos"
		content := container.NewVBox(
			widget.NewLabel("Número de Rodadas:"),
			roundsEntry,
			startButton,
			widget.NewSeparator(),
			outputLabel,
		)

		scroll := container.NewVScroll(content)
		myWindow.SetContent(scroll)
	})

	// Layout da tela inicial
	content := container.NewVBox(
		welcomeLabel,
		normalModeButton,
		allModeButton,
	)
	myWindow.SetContent(container.New(layout.NewCenterLayout(), content))

	// Inicia a aplicação
	myWindow.ShowAndRun()
}
