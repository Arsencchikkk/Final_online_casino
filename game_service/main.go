package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	pb "github.com/Arsencchikkk/final/casino/proto/game"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Card struct {
	Rank string
	Suit string
}

func (c Card) Value() int {
	switch c.Rank {
	case "A":
		return 11
	case "K", "Q", "J":
		return 10
	default:
		v, err := strconv.Atoi(c.Rank)
		if err != nil {
			return 0
		}
		return v
	}
}

// GameSession — состояние одной партии
type GameSession struct {
	Deck       []Card
	PlayerHand []Card
	DealerHand []Card
	State      string
}

var (
	sessions = make(map[string]*GameSession)
	sessMu   sync.RWMutex
)

// Создание и перемешивание колоды юзаем алгоритм
func newDeck() []Card {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	deck := make([]Card, 0, 52)
	for _, s := range suits {
		for _, r := range ranks {
			deck = append(deck, Card{Rank: r, Suit: s})
		}
	}
	rand.Seed(time.Now().UnixNano())
	for i := range deck {
		j := rand.Intn(len(deck))
		deck[i], deck[j] = deck[j], deck[i]
	}
	return deck
}

// Подсчет суммы
func handValue(hand []Card) int {
	total := 0
	aces := 0
	for _, c := range hand {
		total += c.Value()
		if c.Rank == "A" {
			aces++
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

// Создание новой сессии
func newSession() string {
	deck := newDeck()
	player := []Card{deck[0], deck[2]}
	dealer := []Card{deck[1], deck[3]}
	deck = deck[4:]
	session := &GameSession{
		Deck:       deck,
		PlayerHand: player,
		DealerHand: dealer,
		State:      "playerTurn",
	}
	id := uuid.New().String()
	sessMu.Lock()
	sessions[id] = session
	sessMu.Unlock()
	return id
}

func dealerPlay(session *GameSession) {
	for handValue(session.DealerHand) < 17 && len(session.Deck) > 0 {
		session.DealerHand = append(session.DealerHand, session.Deck[0])
		session.Deck = session.Deck[1:]
	}
	session.State = "finished"
}

// Конвертация карты в строку, для логов
func cardToString(c Card) string {
	return c.Rank + c.Suit
}

// gameServer
type gameServer struct {
	pb.UnimplementedGameServiceServer
}

// NewGame:все по новой
func (s *gameServer) NewGame(ctx context.Context, req *pb.NewGameRequest) (*pb.NewGameResponse, error) {
	sessionID := newSession()
	sessMu.RLock()
	session := sessions[sessionID]
	sessMu.RUnlock()

	playerCards := []string{}
	for _, c := range session.PlayerHand {
		playerCards = append(playerCards, cardToString(c))
	}
	dealerCard := cardToString(session.DealerHand[0])
	playerTotal := handValue(session.PlayerHand)

	return &pb.NewGameResponse{
		SessionId:   sessionID,
		PlayerCards: playerCards,
		DealerCard:  dealerCard,
		PlayerTotal: int32(playerTotal),
	}, nil
}

// Hit  игрок берет карту
func (s *gameServer) Hit(ctx context.Context, req *pb.HitRequest) (*pb.HitResponse, error) {
	sessMu.Lock()
	session, ok := sessions[req.SessionId]
	if !ok {
		sessMu.Unlock()
		return nil, fmt.Errorf("session not found")
	}
	if session.State != "playerTurn" {
		sessMu.Unlock()
		return nil, fmt.Errorf("not in player turn")
	}
	if len(session.Deck) > 0 {
		session.PlayerHand = append(session.PlayerHand, session.Deck[0])
		session.Deck = session.Deck[1:]
	}
	playerVal := handValue(session.PlayerHand)
	finished := false
	if playerVal > 21 {
		session.State = "finished"
		finished = true
	}
	sessMu.Unlock()

	// Формируем ответ
	playerCards := []string{}
	for _, c := range session.PlayerHand {
		playerCards = append(playerCards, cardToString(c))
	}
	return &pb.HitResponse{
		PlayerCards: playerCards,
		PlayerTotal: int32(playerVal),
		Finished:    finished,
	}, nil
}

// Stand: игрок останавливаетс дилер добирает
func (s *gameServer) Stand(ctx context.Context, req *pb.StandRequest) (*pb.StandResponse, error) {
	sessMu.Lock()
	session, ok := sessions[req.SessionId]
	if !ok {
		sessMu.Unlock()
		return nil, fmt.Errorf("session not found")
	}
	if session.State != "playerTurn" {
		sessMu.Unlock()
		return nil, fmt.Errorf("not in player turn")
	}
	session.State = "dealerTurn"
	dealerPlay(session)
	dealerVal := handValue(session.DealerHand)

	// Определяем исход
	playerVal := handValue(session.PlayerHand)
	outcome := "draw"
	if playerVal > 21 {
		outcome = "lose"
	} else if dealerVal > 21 {
		outcome = "win"
	} else if playerVal > dealerVal {
		outcome = "win"
	} else if playerVal < dealerVal {
		outcome = "lose"
	}
	sessMu.Unlock()

	dealerCards := []string{}
	for _, c := range session.DealerHand {
		dealerCards = append(dealerCards, cardToString(c))
	}
	return &pb.StandResponse{
		DealerCards: dealerCards,
		DealerTotal: int32(dealerVal),
		Outcome:     outcome,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGameServiceServer(s, &gameServer{})
	log.Println("Blackjack Game Service running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
