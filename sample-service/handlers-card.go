package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func pickACard(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "pickACard")
	defer span.End()

	logger.Info("Picking a card without info context")
	logger.InfoContext(ctx, "Picking a card with info context")

	time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)+50))

	ctx, span = tracer.Start(ctx, "generateDeck")
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(20)+20))
	var deck []Card
	if rand.Intn(3) != 0 {
		deck = generateDeck()
	}
	span.End()

	ctx, span = tracer.Start(ctx, "getRandomCard")
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(20)+20))
	card, err := getRandomCard(deck)
	if err != nil {
		logger.ErrorContext(ctx, "Error picking a card", "error", err)
		http.Error(w, "Error picking a card", http.StatusInternalServerError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return
	}
	span.End()
	logger.InfoContext(ctx, "Picked a card", "card", card)

	span.SetAttributes(attribute.String("card", card.Rank+" of "+card.Suit))

	// ------------------Response Logic-----------------
	resp := fmt.Sprintf("Picked a card: %s of %s\n", card.Rank, card.Suit)
	if _, err := io.WriteString(w, resp); err != nil {
		log.Printf("Write failed: %v\n", err)
	}
	// ------------------------------------------------
}

// Card struct to represent a playing card.
type Card struct {
	Suit string
	Rank string
}

// Function to generate a full deck of cards.
func generateDeck() []Card {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "Jack", "Queen", "King", "Ace"}
	deck := make([]Card, 0, 52) // Pre-allocate for efficiency.

	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

// Function to get a random card from a deck.
func getRandomCard(deck []Card) (Card, error) { //Added error
	if len(deck) == 0 {
		return Card{}, fmt.Errorf("deck is empty") // Return error if the deck is empty
	}
	// Seed the random number generator.  Important for true randomness.
	rand.Seed(time.Now().UnixNano())

	// Generate a random index within the deck's bounds.
	randomIndex := rand.Intn(len(deck)) // No need to check for negative, Intn always returns non-negative

	// Return the card at the random index.
	return deck[randomIndex], nil
}
