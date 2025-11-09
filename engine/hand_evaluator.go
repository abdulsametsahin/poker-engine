package engine

import (
	"poker-engine/models"
	"sort"
)

type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

func (hr HandRank) String() string {
	names := []string{"High Card", "One Pair", "Two Pair", "Three of a Kind", "Straight", "Flush", "Full House", "Four of a Kind", "Straight Flush", "Royal Flush"}
	return names[hr]
}

type HandEvaluation struct {
	Rank    HandRank
	Value   int
	Cards   []models.Card
	Kickers []int
}

func EvaluateHand(playerCards []models.Card, communityCards []models.Card) HandEvaluation {
	allCards := append([]models.Card{}, playerCards...)
	allCards = append(allCards, communityCards...)

	if len(allCards) < 5 {
		return HandEvaluation{Rank: HighCard, Value: 0, Cards: allCards}
	}

	sort.Slice(allCards, func(i, j int) bool {
		return allCards[i].Value() > allCards[j].Value()
	})

	if eval := checkRoyalFlush(allCards); eval.Rank == RoyalFlush {
		return eval
	}
	if eval := checkStraightFlush(allCards); eval.Rank == StraightFlush {
		return eval
	}
	if eval := checkFourOfAKind(allCards); eval.Rank == FourOfAKind {
		return eval
	}
	if eval := checkFullHouse(allCards); eval.Rank == FullHouse {
		return eval
	}
	if eval := checkFlush(allCards); eval.Rank == Flush {
		return eval
	}
	if eval := checkStraight(allCards); eval.Rank == Straight {
		return eval
	}
	if eval := checkThreeOfAKind(allCards); eval.Rank == ThreeOfAKind {
		return eval
	}
	if eval := checkTwoPair(allCards); eval.Rank == TwoPair {
		return eval
	}
	if eval := checkOnePair(allCards); eval.Rank == OnePair {
		return eval
	}

	return checkHighCard(allCards)
}

func CompareHands(eval1, eval2 HandEvaluation) int {
	if eval1.Value > eval2.Value {
		return 1
	}
	if eval1.Value < eval2.Value {
		return -1
	}
	return 0
}

func checkRoyalFlush(cards []models.Card) HandEvaluation {
	eval := checkStraightFlush(cards)
	if eval.Rank == StraightFlush && len(eval.Cards) > 0 && eval.Cards[0].Value() == 14 {
		return HandEvaluation{Rank: RoyalFlush, Value: 100000, Cards: eval.Cards}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkStraightFlush(cards []models.Card) HandEvaluation {
	suitMap := make(map[models.Suit][]models.Card)
	for _, card := range cards {
		suitMap[card.Suit] = append(suitMap[card.Suit], card)
	}

	for _, suitCards := range suitMap {
		if len(suitCards) >= 5 {
			straight := findStraight(suitCards)
			if len(straight) >= 5 {
				return HandEvaluation{Rank: StraightFlush, Value: 90000 + straight[0].Value(), Cards: straight[:5]}
			}
		}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkFourOfAKind(cards []models.Card) HandEvaluation {
	rankCount := make(map[models.Rank][]models.Card)
	for _, card := range cards {
		rankCount[card.Rank] = append(rankCount[card.Rank], card)
	}

	for rank, rankCards := range rankCount {
		if len(rankCards) == 4 {
			var kicker models.Card
			for _, card := range cards {
				if card.Rank != rank && (kicker.Rank == "" || card.Value() > kicker.Value()) {
					kicker = card
				}
			}
			bestCards := append(rankCards, kicker)
			return HandEvaluation{Rank: FourOfAKind, Value: 80000 + rankCards[0].Value()*100 + kicker.Value(), Cards: bestCards[:5]}
		}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkFullHouse(cards []models.Card) HandEvaluation {
	rankCount := make(map[models.Rank][]models.Card)
	for _, card := range cards {
		rankCount[card.Rank] = append(rankCount[card.Rank], card)
	}

	var threeCards, pairCards []models.Card
	var bestThreeValue int

	// Find the best three of a kind
	for _, rankCards := range rankCount {
		if len(rankCards) >= 3 {
			if len(threeCards) == 0 || rankCards[0].Value() > bestThreeValue {
				threeCards = rankCards[:3]
				bestThreeValue = rankCards[0].Value()
			}
		}
	}

	// Find the best pair (different from the three of a kind)
	for _, rankCards := range rankCount {
		if len(rankCards) >= 2 && len(threeCards) > 0 && rankCards[0].Rank != threeCards[0].Rank {
			if len(pairCards) == 0 || rankCards[0].Value() > pairCards[0].Value() {
				pairCards = rankCards[:2]
			}
		}
	}

	if len(threeCards) > 0 && len(pairCards) > 0 {
		bestCards := append(threeCards, pairCards...)
		return HandEvaluation{Rank: FullHouse, Value: 70000 + threeCards[0].Value()*100 + pairCards[0].Value(), Cards: bestCards}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkFlush(cards []models.Card) HandEvaluation {
	suitMap := make(map[models.Suit][]models.Card)
	for _, card := range cards {
		suitMap[card.Suit] = append(suitMap[card.Suit], card)
	}

	for _, suitCards := range suitMap {
		if len(suitCards) >= 5 {
			sort.Slice(suitCards, func(i, j int) bool {
				return suitCards[i].Value() > suitCards[j].Value()
			})
			value := 60000
			for i := 0; i < 5; i++ {
				value += suitCards[i].Value() * (1 << (4 - i))
			}
			return HandEvaluation{Rank: Flush, Value: value, Cards: suitCards[:5]}
		}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkStraight(cards []models.Card) HandEvaluation {
	straight := findStraight(cards)
	if len(straight) >= 5 {
		return HandEvaluation{Rank: Straight, Value: 50000 + straight[0].Value(), Cards: straight[:5]}
	}
	return HandEvaluation{Rank: HighCard}
}

func findStraight(cards []models.Card) []models.Card {
	uniqueRanks := make(map[int]models.Card)
	for _, card := range cards {
		val := card.Value()
		if _, exists := uniqueRanks[val]; !exists {
			uniqueRanks[val] = card
		}
	}

	values := make([]int, 0, len(uniqueRanks))
	for val := range uniqueRanks {
		values = append(values, val)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(values)))

	consecutive := []models.Card{uniqueRanks[values[0]]}
	for i := 1; i < len(values); i++ {
		if values[i-1]-values[i] == 1 {
			consecutive = append(consecutive, uniqueRanks[values[i]])
			if len(consecutive) >= 5 {
				return consecutive
			}
		} else {
			consecutive = []models.Card{uniqueRanks[values[i]]}
		}
	}

	// Check for wheel (A-2-3-4-5) - Ace acts as low card
	if len(values) >= 5 && values[0] == 14 {
		// Check if we have 5, 4, 3, 2
		hasWheel := true
		wheel := []models.Card{}
		for _, val := range []int{5, 4, 3, 2} {
			if card, exists := uniqueRanks[val]; exists {
				wheel = append(wheel, card)
			} else {
				hasWheel = false
				break
			}
		}
		if hasWheel && len(wheel) == 4 {
			// Add the Ace at the end (acts as low card)
			wheel = append(wheel, uniqueRanks[14])
			return wheel
		}
	}

	return []models.Card{}
}

func checkThreeOfAKind(cards []models.Card) HandEvaluation {
	rankCount := make(map[models.Rank][]models.Card)
	for _, card := range cards {
		rankCount[card.Rank] = append(rankCount[card.Rank], card)
	}

	for _, rankCards := range rankCount {
		if len(rankCards) >= 3 {
			kickers := []models.Card{}
			for _, card := range cards {
				if card.Rank != rankCards[0].Rank {
					kickers = append(kickers, card)
				}
			}
			sort.Slice(kickers, func(i, j int) bool {
				return kickers[i].Value() > kickers[j].Value()
			})

			// Safety check for kickers
			if len(kickers) < 2 {
				// Should not happen with 7 cards, but handle gracefully
				bestCards := rankCards[:3]
				bestCards = append(bestCards, kickers...)
				value := 40000 + rankCards[0].Value()*100
				if len(kickers) > 0 {
					value += kickers[0].Value() * 10
				}
				return HandEvaluation{Rank: ThreeOfAKind, Value: value, Cards: bestCards}
			}

			bestCards := append(rankCards[:3], kickers[:2]...)
			value := 40000 + rankCards[0].Value()*100 + kickers[0].Value()*10 + kickers[1].Value()
			return HandEvaluation{Rank: ThreeOfAKind, Value: value, Cards: bestCards[:5]}
		}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkTwoPair(cards []models.Card) HandEvaluation {
	rankCount := make(map[models.Rank][]models.Card)
	for _, card := range cards {
		rankCount[card.Rank] = append(rankCount[card.Rank], card)
	}

	pairs := [][]models.Card{}
	for _, rankCards := range rankCount {
		if len(rankCards) >= 2 {
			pairs = append(pairs, rankCards[:2])
		}
	}

	if len(pairs) >= 2 {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i][0].Value() > pairs[j][0].Value()
		})

		var kicker models.Card
		for _, card := range cards {
			if card.Rank != pairs[0][0].Rank && card.Rank != pairs[1][0].Rank {
				if kicker.Rank == "" || card.Value() > kicker.Value() {
					kicker = card
				}
			}
		}

		bestCards := append(append(pairs[0], pairs[1]...), kicker)
		value := 30000 + pairs[0][0].Value()*100 + pairs[1][0].Value()*10 + kicker.Value()
		return HandEvaluation{Rank: TwoPair, Value: value, Cards: bestCards[:5]}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkOnePair(cards []models.Card) HandEvaluation {
	rankCount := make(map[models.Rank][]models.Card)
	for _, card := range cards {
		rankCount[card.Rank] = append(rankCount[card.Rank], card)
	}

	for _, rankCards := range rankCount {
		if len(rankCards) >= 2 {
			kickers := []models.Card{}
			for _, card := range cards {
				if card.Rank != rankCards[0].Rank {
					kickers = append(kickers, card)
				}
			}
			sort.Slice(kickers, func(i, j int) bool {
				return kickers[i].Value() > kickers[j].Value()
			})

			// Safety check for kickers
			if len(kickers) < 3 {
				// Should not happen with 7 cards, but handle gracefully
				bestCards := rankCards[:2]
				bestCards = append(bestCards, kickers...)
				value := 20000 + rankCards[0].Value()*1000
				for i, k := range kickers {
					if i < 3 {
						value += k.Value() * (100 / (i + 1))
					}
				}
				return HandEvaluation{Rank: OnePair, Value: value, Cards: bestCards}
			}

			bestCards := append(rankCards[:2], kickers[:3]...)
			value := 20000 + rankCards[0].Value()*1000 + kickers[0].Value()*100 + kickers[1].Value()*10 + kickers[2].Value()
			return HandEvaluation{Rank: OnePair, Value: value, Cards: bestCards[:5]}
		}
	}
	return HandEvaluation{Rank: HighCard}
}

func checkHighCard(cards []models.Card) HandEvaluation {
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Value() > cards[j].Value()
	})

	value := 10000
	for i := 0; i < 5 && i < len(cards); i++ {
		value += cards[i].Value() * (1 << (4 - i))
	}

	bestCards := cards
	if len(bestCards) > 5 {
		bestCards = bestCards[:5]
	}

	return HandEvaluation{Rank: HighCard, Value: value, Cards: bestCards}
}
