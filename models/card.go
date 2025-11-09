package models

import (
	"fmt"
	"math/rand"
	"time"
)

type Suit string
type Rank string

const (
	Hearts   Suit = "h"
	Diamonds Suit = "d"
	Clubs    Suit = "c"
	Spades   Suit = "s"
)

const (
	Two   Rank = "2"
	Three Rank = "3"
	Four  Rank = "4"
	Five  Rank = "5"
	Six   Rank = "6"
	Seven Rank = "7"
	Eight Rank = "8"
	Nine  Rank = "9"
	Ten   Rank = "T"
	Jack  Rank = "J"
	Queen Rank = "Q"
	King  Rank = "K"
	Ace   Rank = "A"
)

type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit)
}

func (c Card) Value() int {
	switch c.Rank {
	case Two:
		return 2
	case Three:
		return 3
	case Four:
		return 4
	case Five:
		return 5
	case Six:
		return 6
	case Seven:
		return 7
	case Eight:
		return 8
	case Nine:
		return 9
	case Ten:
		return 10
	case Jack:
		return 11
	case Queen:
		return 12
	case King:
		return 13
	case Ace:
		return 14
	}
	return 0
}

type Deck struct {
	cards []Card
	rng   *rand.Rand
}

func NewDeck() *Deck {
	deck := &Deck{
		cards: make([]Card, 0, 52),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	deck.Reset()
	return deck
}

func (d *Deck) Reset() {
	d.cards = make([]Card, 0, 52)
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	ranks := []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}

	for _, suit := range suits {
		for _, rank := range ranks {
			d.cards = append(d.cards, Card{Rank: rank, Suit: suit})
		}
	}
	d.Shuffle()
}

func (d *Deck) Shuffle() {
	d.rng.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

func (d *Deck) Deal() (Card, error) {
	if len(d.cards) == 0 {
		return Card{}, fmt.Errorf("deck is empty - no more cards to deal")
	}
	card := d.cards[0]
	d.cards = d.cards[1:]
	return card, nil
}

func (d *Deck) DealMultiple(n int) ([]Card, error) {
	if len(d.cards) < n {
		return nil, fmt.Errorf("not enough cards in deck: requested %d, available %d", n, len(d.cards))
	}
	cards := make([]Card, n)
	for i := 0; i < n; i++ {
		card, err := d.Deal()
		if err != nil {
			return nil, err
		}
		cards[i] = card
	}
	return cards, nil
}

func (d *Deck) CardsRemaining() int {
	return len(d.cards)
}
