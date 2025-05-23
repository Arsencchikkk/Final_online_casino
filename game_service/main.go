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
		v, _ := strconv.Atoi(c.Rank)
		return v
	}
}

func cardToString(c Card) string {
	return c.Rank + c.Suit
}

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

func handValue(hand []Card) int {
	sum, aces := 0, 0
	for _, c := range hand {
		sum += c.Value()
		if c.Rank == "A" {
			aces++
		}
	}
	for sum > 21 && aces > 0 {
		sum -= 10
		aces--
	}
	return sum
}

type GameSession struct {
	Deck       []Card
	PlayerHand []Card
	DealerHand []Card
	State      string // "playerTurn", "dealerTurn", "finished"
}

var (
	sessions = make(map[string]*GameSession)
	sessMu   sync.RWMutex
)

func newSession() string {
	d := newDeck()
	// deal: player 0, dealer 1, player 2, dealer 3
	player := []Card{d[0], d[2]}
	dealer := []Card{d[1], d[3]}
	s := &GameSession{
		Deck:       d[4:],
		PlayerHand: player,
		DealerHand: dealer,
		State:      "playerTurn",
	}
	id := uuid.New().String()
	sessMu.Lock()
	sessions[id] = s
	sessMu.Unlock()
	return id
}

func dealerPlay(s *GameSession) {
	for handValue(s.DealerHand) < 17 && len(s.Deck) > 0 {
		s.DealerHand = append(s.DealerHand, s.Deck[0])
		s.Deck = s.Deck[1:]
	}
	s.State = "finished"
}

type gameServer struct {
	pb.UnimplementedGameServiceServer
}

func (s *gameServer) NewGame(ctx context.Context, _ *pb.NewGameRequest) (*pb.NewGameResponse, error) {
	id := newSession()

	sessMu.RLock()
	session := sessions[id]
	sessMu.RUnlock()

	// player cards
	pc := make([]string, len(session.PlayerHand))
	for i, c := range session.PlayerHand {
		pc[i] = cardToString(c)
	}
	// dealer cards (both, from the start)
	dc := make([]string, len(session.DealerHand))
	for i, c := range session.DealerHand {
		dc[i] = cardToString(c)
	}

	return &pb.NewGameResponse{
		SessionId:   id,
		PlayerCards: pc,
		DealerCards: dc,
		PlayerTotal: int32(handValue(session.PlayerHand)),
		Balance:     0, // stub; plug in wallet call if you like
	}, nil
}

func (s *gameServer) Hit(ctx context.Context, req *pb.HitRequest) (*pb.HitResponse, error) {
	sessMu.Lock()
	defer sessMu.Unlock()

	session, ok := sessions[req.SessionId]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	// Only allow hits while in “playerTurn”
	if session.State == "playerTurn" && len(session.Deck) > 0 {
		session.PlayerHand = append(session.PlayerHand, session.Deck[0])
		session.Deck = session.Deck[1:]
	}

	val := handValue(session.PlayerHand)
	finished := false
	if val > 21 {
		session.State = "finished"
		finished = true
	}

	pc := make([]string, len(session.PlayerHand))
	for i, c := range session.PlayerHand {
		pc[i] = cardToString(c)
	}

	return &pb.HitResponse{
		PlayerCards: pc,
		PlayerTotal: int32(val),
		Finished:    finished,
		Balance:     0,
	}, nil
}

func (s *gameServer) Stand(ctx context.Context, req *pb.StandRequest) (*pb.StandResponse, error) {
	sessMu.Lock()
	session, ok := sessions[req.SessionId]
	if !ok {
		sessMu.Unlock()
		return nil, fmt.Errorf("session not found")
	}
	// If still in player turn, run dealer
	if session.State == "playerTurn" {
		session.State = "dealerTurn"
		dealerPlay(session)
	}

	// dealer cards & totals
	dc := make([]string, len(session.DealerHand))
	for i, c := range session.DealerHand {
		dc[i] = cardToString(c)
	}
	dTotal := handValue(session.DealerHand)
	pTotal := handValue(session.PlayerHand)

	// decide outcome
	outcome := "push"
	switch {
	case pTotal > 21:
		outcome = "lose"
	case dTotal > 21 || pTotal > dTotal:
		outcome = "win"
	case pTotal < dTotal:
		outcome = "lose"
	}

	sessMu.Unlock()
	return &pb.StandResponse{
		DealerCards: dc,
		DealerTotal: int32(dTotal),
		Outcome:     outcome,
		Balance:     0,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterGameServiceServer(srv, &gameServer{})
	log.Println("Game Service listening on :50051")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve error: %v", err)
	}
}
