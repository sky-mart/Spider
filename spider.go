package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const (
	SuitCount      = 4
	RankCount      = 13
	DeckSize       = SuitCount * RankCount
	SpiderDeckSize = 2 * DeckSize
	SpiderColCount = 10
)

type GameResult int
const (
    YouLose GameResult = iota
    YouWin
    NoResult
)

var suits = [SuitCount]string{
	"♥️", "♦️", "️♠️", "♣️",
}
var ranks = [RankCount]string{
	"A", "2", "3", "4", "5", "6", "7",
	"8", "9", "10", "J", "Q", "K",
}

type Card struct {
    Suit   int  `json:"suit"`
	Rank   int  `json:"rank"`
	Faceup bool `json:"faceup"`
}

type SpiderMove struct {
	From      int `json:"from"`
	To        int `json:"to"`
	CardCount int `json:"cardcount"`
    Opening bool
    SuitCleared int
}

type SpiderState struct {
	Field   [SpiderColCount][]Card
	Deck    []Card
	History []*SpiderMove
}

func (card Card) str() string {
	if !card.Faceup {
		return "*"
	}
	return ranks[card.Rank] + " " + suits[card.Suit]
}

func removeAtIndex(source []int, index int) []int {
	lastIndex := len(source) - 1
	//swap the last value and the value we want to remove
	source[index], source[lastIndex] = source[lastIndex], source[index]
	return source[:lastIndex]
}

func randSeq(size int) []int {
	workseq := make([]int, size)
	for i := 0; i < size; i++ {
		workseq[i] = i
	}
	seq := make([]int, 0, size)
	for i := 0; i < size; i++ {
		j := rand.Intn(len(workseq))
		seq = append(seq, workseq[j])
		workseq = removeAtIndex(workseq, j)
	}
	return seq
}

func createDeck(shuffled bool) [DeckSize]Card {
	var deck [DeckSize]Card
	var seq []int
	var cardIndex, seqIndex int

	if shuffled {
		seq = randSeq(DeckSize)
		seqIndex = 0
	} else {
		cardIndex = -1
	}

	for suit := 0; suit < SuitCount; suit++ {
		for rank := 0; rank < RankCount; rank++ {
			if shuffled {
				cardIndex = seq[seqIndex]
				seqIndex++
			} else {
				cardIndex++
			}
			deck[cardIndex] = Card{Suit: suit, Rank: rank}
		}
	}
	return deck
}

func createSpiderDeck() []Card {
	deck := make([]Card, SpiderDeckSize, SpiderDeckSize)
	tmpDeck := createDeck(true)
	for i := 0; i < DeckSize; i++ {
		deck[i] = tmpDeck[i]
	}
	tmpDeck = createDeck(true)
	for i := 0; i < DeckSize; i++ {
		deck[DeckSize+i] = tmpDeck[i]
	}
	return deck
}

func initSpiderState() (state *SpiderState) {
	state = &SpiderState{
		Deck: createSpiderDeck(),
	}
	for row := 0; row < 4; row++ {
		for col := 0; col < SpiderColCount; col++ {
			state.Field[col] = append(state.Field[col],
				state.Deck[len(state.Deck)-1])
			state.Deck = state.Deck[:len(state.Deck)-1]
		}
	}
	for col := 0; col < 4; col++ {
		state.Field[col] = append(state.Field[col],
			state.Deck[len(state.Deck)-1])
		state.Deck = state.Deck[:len(state.Deck)-1]
	}
	for col := 4; col < SpiderColCount; col++ {
		coming := state.Deck[len(state.Deck)-1]
		coming.Faceup = true
		state.Field[col] = append(state.Field[col], coming)
		state.Deck = state.Deck[:len(state.Deck)-1]
	}
	for col := 0; col < 4; col++ {
		coming := state.Deck[len(state.Deck)-1]
		coming.Faceup = true
		state.Field[col] = append(state.Field[col], coming)
		state.Deck = state.Deck[:len(state.Deck)-1]
	}
	return state
}

func (state *SpiderState) str() string {
	var s string
	for row, allColsFinished := 0, false; !allColsFinished; row++ {
		allColsFinished = true
		for col := 0; col < SpiderColCount; col++ {
			if row < len(state.Field[col]) {
				allColsFinished = false
				s += state.Field[col][row].str() + "\t"
			} else {
                if row == 0 {
                    s += "-"
                    s += "-"
                }
				s += "\t"
			}
		}
		s += "\n"
	}
	return s
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (state *SpiderState) writeTransparentSave(savePath string) {
	f, err := os.Create(savePath)
	check(err)
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(state)
	check(err)
}

func readTransparentSave(savePath string) (state *SpiderState) {
	f, err := os.Open(savePath)
	check(err)
	defer f.Close()

	state = &SpiderState{}
	dec := json.NewDecoder(f)
	err = dec.Decode(state)
	check(err)
	return state
}

func (state *SpiderState) isMovePossible(move *SpiderMove) bool {
	// is it possible to take cards from
	fromLen := len(state.Field[move.From])
	if fromLen < move.CardCount {
		return false
	}
	fromChain := state.Field[move.From][fromLen-move.CardCount:]
	//fmt.Println(fromChain)
	for i := move.CardCount - 1; i >= 0; i-- {
		if fromChain[i].Faceup == false {
			return false
		}
		if i > 0 {
			sameSuit := fromChain[i].Suit == fromChain[i-1].Suit
			incRank := fromChain[i].Rank+1 == fromChain[i-1].Rank
			if !sameSuit || !incRank {
				return false
			}
		}
	}
	// is it possible to put cards to
	toCol := state.Field[move.To]
	//fmt.Println(fromChain[0].Rank)
	//fmt.Println(toCol[len(toCol)-1].Rank)
	return (len(toCol) == 0) || (fromChain[0].Rank+1 == toCol[len(toCol)-1].Rank)
}

func (state *SpiderState) makeMove(move *SpiderMove, fwd bool) {
	if fwd {
        fromCol := state.Field[move.From]
        toCol := state.Field[move.To]
        fromLen := len(fromCol)
        fromChain := fromCol[fromLen-move.CardCount:]
        state.Field[move.From] = fromCol[:fromLen-move.CardCount]
        state.Field[move.To] = append(toCol, fromChain...)

        fromCol = state.Field[move.From]
        fromLen = len(fromCol)
        if fromLen > 0 {
            if fromCol[fromLen-1].Faceup == false {
                move.Opening = true
                fromCol[fromLen-1].Faceup = true
            }
        }

        // clear suit chain
        if fromChain[len(fromChain)-1].Rank != 0 {
             return 
        }
        toCol = state.Field[move.To]
        toLen := len(toCol)
        for i := toLen - len(fromChain) - 1; i >= toLen - RankCount; i-- {
            sameSuit := toCol[i].Suit == toCol[i+1].Suit
            decRank := toCol[i].Rank - 1 == toCol[i+1].Rank
            if !sameSuit || !decRank {
                return 
            }
        }
        move.SuitCleared = toCol[toLen-1].Suit
        state.Field[move.To] = toCol[:toLen - RankCount]
	} else {
        if move.SuitCleared != -1 {
            chain := make([]Card, RankCount)
            for i := 0; i < RankCount; i++ {
                chain[i] = Card{
                    Suit: move.SuitCleared,
                    Rank: RankCount - 1 - i,
                    Faceup: true,
                }
            }
            state.Field[move.From] = append(state.Field[move.From], chain...)
            move.SuitCleared = -1
            state.makeMove(move, false)
            return 
        }
        fromCol := state.Field[move.From]
        toCol := state.Field[move.To]
        fromLen := len(fromCol)
        fromChain := fromCol[fromLen-move.CardCount:]
        state.Field[move.From] = fromCol[:fromLen-move.CardCount]
        state.Field[move.To] = append(toCol, fromChain...)

        if move.Opening {
	    	state.Field[move.To][len(toCol)-1].Faceup = false
        }
	}
}

func (state *SpiderState) doMove(move *SpiderMove) (bool, GameResult ) {
	if state.isMovePossible(move) {
		state.makeMove(move, true)
		state.History = append(state.History, move)
        if move.SuitCleared != -1 {
            for i := 0; i < SpiderColCount; i++ {
                if len(state.Field[i]) > 0 {
                    fmt.Printf("len != 0, %d", i)
                    return true, NoResult
                }
            }
            return true, YouWin
        }
        if len(state.possibleMoves()) == 0 {
            return true, YouLose
        }
		return true, NoResult
	}
	return false, NoResult
}

func (state *SpiderState) undoLastMove() bool {
	if len(state.History) == 0 {
		return false
	}
	lastMove := state.History[len(state.History)-1]
	if lastMove.CardCount == 0 {
		for col := SpiderColCount - 1; col >= 0; col-- {
			state.Field[col][len(state.Field[col])-1].Faceup = false
			state.Deck = append(state.Deck,
				state.Field[col][len(state.Field[col])-1])
			state.Field[col] = state.Field[col][:len(state.Field[col])-1]
		}
	} else {
		lastMove.From, lastMove.To = lastMove.To, lastMove.From
		state.makeMove(lastMove, false)
	}
	state.History = state.History[:len(state.History)-1]
	return true
}

func (state *SpiderState) newRow() bool {
	if len(state.Deck) > 0 {
		for col := 0; col < SpiderColCount; col++ {
			state.Field[col] = append(state.Field[col],
				state.Deck[len(state.Deck)-1-col])
			state.Field[col][len(state.Field[col])-1].Faceup = true
		}
		state.Deck = state.Deck[:len(state.Deck)-SpiderColCount]
		state.History = append(state.History,
			&SpiderMove{CardCount: 0})
		return true
	}
	return false
}

func readMove(s string) (*SpiderMove, error) {
	move := &SpiderMove{}
	var err error
	ms := strings.Split(s, "-")
	if len(ms) != 3 {
		err = errors.New("Wrong number of arguments")
		return move, err
	}
	move.From, err = strconv.Atoi(ms[0])
	if err != nil {
		return move, err
	}
	move.To, err = strconv.Atoi(ms[1])
	if err != nil {
		return move, err
	}
	move.CardCount, err = strconv.Atoi(ms[2])
	return move, err
}

func (state *SpiderState) possibleMoves() []*SpiderMove {
    moves := []*SpiderMove{}
    for from := 0; from < SpiderColCount; from++ {
        for to := 0; to < SpiderColCount; to++ {
            if to == from {
                continue
            }
            for count := 1; ; count++ {
                move := &SpiderMove{from, to, count, false, -1}
                if state.isMovePossible(move) {
                    moves = append(moves, move)
                } else {
                    break;
                }
            }
        }
    }
    return moves
}

// TODO:
// ---1) Создать репозиторий для отслеживания прогресса
// ---2) Функции записи/чтения в файл состояния игры
// ---3) Ход. Соответствие хода правилам
// ---4) История. Возврат ходов
// ---5) Сдача карт из колоды
// ---6) Консольная версия: считать ход, отмену хода, сдачу карт, сейв
// 7) Дерево вариантов

func main() {
	var state *SpiderState
	var ans string

	fmt.Println("Hello! This is Spider Solitaire.\nWould you like to start a new game or load an old one? (n/l):")

	fmt.Scanf("%s", &ans)
	switch ans {
	case "n", "new":
		fmt.Println("Starting a new game.")
		state = initSpiderState()
	case "l", "load":
		fmt.Println("Loading game. Choose save: (filepath)")
		fmt.Scanf("%s", &ans)
		state = readTransparentSave(ans)
	}
	fmt.Print(state.str())

	for {
		fmt.Scanf("%s", &ans)
		switch ans {
		case "q", "quit":
			return
		case "s", "save":
			fmt.Println("Saving game. Choose save: (filepath)")
			fmt.Scanf("%s", &ans)
			state.writeTransparentSave(ans)
			return
		case "r", "row":
			if state.newRow() {
				fmt.Print(state.str())
			} else {
				fmt.Println("No cards left in the deck")
			}
		case "u", "undo":
			if state.undoLastMove() {
				fmt.Print(state.str())
			} else {
				fmt.Println("No history")
			}
        case "p", "possible":
            moves := state.possibleMoves()
            for _, m := range moves {
                fmt.Println(m)
            }
		default:
			move, err := readMove(ans)
			if err != nil {
				fmt.Println("Wrong move format. Use <from-to-count>")
				break
			}
            ok, res := state.doMove(move)
			if ok {
                fmt.Print(state.str())
                if res == YouWin {
                    fmt.Println("You win!")
                } else if res == YouLose {
                    fmt.Println("You lose!")
                }
			} else {
				fmt.Println("Impossible move")
			}
		}
	}
}
