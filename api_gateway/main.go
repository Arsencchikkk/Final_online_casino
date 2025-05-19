package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	gamepb "github.com/Arsencchikkk/final/casino/proto/game"
	walletpb "github.com/Arsencchikkk/final/casino/proto/wallet"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// corsMiddleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {

	gameConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to game service: %v", err)
	}
	defer gameConn.Close()
	gameClient := gamepb.NewGameServiceClient(gameConn)

	walletConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to wallet service: %v", err)
	}
	defer walletConn.Close()
	walletClient := walletpb.NewWalletServiceClient(walletConn)

	http.HandleFunc("/new_game", func(w http.ResponseWriter, r *http.Request) {
		resp, err := gameClient.NewGame(context.Background(), &gamepb.NewGameRequest{})
		if err != nil {
			http.Error(w, "NewGame error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/hit", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			http.Error(w, "Missing session_id", http.StatusBadRequest)
			return
		}
		resp, err := gameClient.Hit(context.Background(), &gamepb.HitRequest{SessionId: sessionID})
		if err != nil {
			http.Error(w, "Hit error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/stand", func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			http.Error(w, "Missing session_id", http.StatusBadRequest)
			return
		}

		standResp, err := gameClient.Stand(context.Background(), &gamepb.StandRequest{SessionId: sessionID})
		if err != nil {
			http.Error(w, "Stand error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		bet := 100
		var netChange int32
		switch standResp.Outcome {
		case "win":
			netChange = int32(bet * 2)
		case "lose":
			netChange = -int32(bet)
		default:
			netChange = 0
		}

		walletResp, err := walletClient.UpdateBalance(context.Background(), &walletpb.WalletUpdateRequest{
			UserId: "testuser",
			Amount: netChange,
		})
		if err != nil {
			http.Error(w, "Wallet update error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		combinedResp := map[string]interface{}{
			"game":   standResp,
			"wallet": walletResp,
		}
		json.NewEncoder(w).Encode(combinedResp)
	})

	// Эндпоинт
	http.HandleFunc("/wallet_balance", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "Missing user_id", http.StatusBadRequest)
			return
		}
		resp, err := walletClient.GetBalance(context.Background(), &walletpb.WalletRequest{UserId: userID})
		if err != nil {
			http.Error(w, "Wallet balance error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

	handler := corsMiddleware(http.DefaultServeMux)
	log.Println("API Gateway running on port 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
