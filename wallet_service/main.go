package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/Arsencchikkk/final/casino/proto/wallet"
	"github.com/joho/godotenv"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"

	"google.golang.org/grpc"
)

// walletServer
type walletServer struct {
	pb.UnimplementedWalletServiceServer
	db  *sql.DB
	rdb *redis.Client
}

// NewWalletServer
func NewWalletServer(ctx context.Context) *walletServer {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}

	connStr := os.Getenv("DB_DSN")
	if connStr == "" {
		log.Fatal("DB_DSN not set in environment")
	}

	// Инициализация PostgreSQL.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	// Проверка бд.
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping postgres: %v", err)
	}

	// Инициализация Redis.
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	// Проверяем соединение с Redis.
	if err = rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to ping redis: %v", err)
	}

	return &walletServer{
		db:  db,
		rdb: rdb,
	}
}

func (s *walletServer) GetBalance(ctx context.Context, req *pb.WalletRequest) (*pb.WalletResponse, error) {
	// Пытаемся получить баланс из Redis.
	balanceStr, err := s.rdb.Get(ctx, req.UserId).Result()
	if err == nil {
		var balance int
		if _, err := fmt.Sscanf(balanceStr, "%d", &balance); err == nil {
			return &pb.WalletResponse{Balance: int32(balance)}, nil
		}
	}

	// Чтение из PostgreSQL.
	var balance int
	err = s.db.QueryRowContext(ctx, "SELECT balance FROM wallets WHERE user_id = $1", req.UserId).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если записи нет, возвращаем 0.
			return &pb.WalletResponse{Balance: 0}, nil
		}
		return nil, fmt.Errorf("failed to query balance: %w", err)
	}

	// Кэшируем значение в Redis.
	if err := s.rdb.Set(ctx, req.UserId, balance, 0).Err(); err != nil {
		log.Printf("failed to cache balance for user %s: %v", req.UserId, err)
	}
	return &pb.WalletResponse{Balance: int32(balance)}, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	ws := NewWalletServer(ctx)
	pb.RegisterWalletServiceServer(s, ws)

	log.Println("Wallet service running on port 50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
