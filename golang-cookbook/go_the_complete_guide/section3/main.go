package main

// func newCard() string {
// 	return "Five of Diamonds"
// }

func main() {
	cards := newDeck()
	cards.shuffle()
	cards.print()
}
