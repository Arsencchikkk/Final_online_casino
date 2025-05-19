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

// corsMiddleware оборачивает API, чтобы разрешить CORS
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
	// 1) Подключаемся к gRPC-сервисам
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

	// 2) Собираем mux
	mux := http.NewServeMux()

	// 2.1) Регистрируем все API под /api/
	api := http.NewServeMux()
	api.HandleFunc("/new_game", newGameHandler(gameClient))
	api.HandleFunc("/hit", hitHandler(gameClient))
	api.HandleFunc("/stand", standHandler(gameClient, walletClient))
	api.HandleFunc("/wallet", balanceHandler(walletClient))

	// Вешаем CORS и StripPrefix
	mux.Handle("/api/", corsMiddleware(http.StripPrefix("/api", api)))

	// 2.2) Всё остальное — это фронтенд (HTML/JS/CSS) из папки Front
	fs := http.FileServer(http.Dir("../Front"))
	mux.Handle("/", fs)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// --- handlers ---

func newGameHandler(gameClient gamepb.GameServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := gameClient.NewGame(context.Background(), &gamepb.NewGameRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func hitHandler(gameClient gamepb.GameServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		resp, err := gameClient.Hit(context.Background(), &gamepb.HitRequest{SessionId: sessionID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func standHandler(gameClient gamepb.GameServiceClient, walletClient walletpb.WalletServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		standResp, err := gameClient.Stand(context.Background(), &gamepb.StandRequest{SessionId: sessionID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// жизненный цикл пари: win → +2×ставка, lose → −1×ставка
		bet := int32(100)
		var change int32
		switch standResp.Outcome {
		case "win":
			change = bet * 2
		case "lose":
			change = -bet
		}
		walletResp, err := walletClient.UpdateBalance(context.Background(), &walletpb.WalletUpdateRequest{
			UserId: "testuser", Amount: change,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"game":   standResp,
			"wallet": walletResp,
		})
	}
}

func balanceHandler(walletClient walletpb.WalletServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		resp, err := walletClient.GetBalance(context.Background(), &walletpb.WalletRequest{UserId: userID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	}
}
