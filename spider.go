package main

import (
	"fmt"
	"math/rand"
)

const (
	SuitCount      = 4
	RankCount      = 13
	DeckSize       = SuitCount * RankCount
	SpiderDeckSize = 2 * DeckSize
	SpiderColCount = 10
)

var suits = [SuitCount]string{
	"♥️", "♦️", "️♠️", "♣️",
}
var ranks = [RankCount]string{
	"A", "2", "3", "4", "5", "6", "7",
	"8", "9", "10", "J", "Q", "K",
}

type Card struct {
	Suit   int
	Rank   int
	Faceup bool
}

type SpiderState struct {
	Field [SpiderColCount][]Card
	Deck  []Card
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
				s += "\t"
			}
		}
		s += "\n"
	}
	return s
}

// TODO:
// 1) Создать репозиторий для отслеживания прогресса
// 2) Функции записи/чтения в файл состояния игры
// 3) Ход. Соответствие хода правилам
// 4) Дерево вариантов

func main() {
	fmt.Println("Hello, Spider!")

	state := initSpiderState()
	fmt.Print(state.str())
}
