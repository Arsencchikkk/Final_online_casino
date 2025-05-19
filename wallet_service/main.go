package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/Arsencchikkk/final/casino/proto/wallet"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type walletServer struct {
	pb.UnimplementedWalletServiceServer
	col *mongo.Collection
}

// NewWalletServer инициализирует подключение к MongoDB
func NewWalletServer(ctx context.Context) *walletServer {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or failed to load: %v", err)
	}
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	colName := os.Getenv("MONGO_COLLECTION")
	if uri == "" || dbName == "" || colName == "" {
		log.Fatal("MONGO_URI, MONGO_DB or MONGO_COLLECTION not set in environment")
	}

	// Подключаемся к MongoDB Atlas
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}
	col := client.Database(dbName).Collection(colName)
	return &walletServer{col: col}
}

// GetBalance возвращает баланс пользователя (0, если нет записи)
func (s *walletServer) GetBalance(ctx context.Context, req *pb.WalletRequest) (*pb.WalletResponse, error) {
	var doc struct {
		UserID  string `bson:"user_id"`
		Balance int32  `bson:"balance"`
	}
	err := s.col.FindOne(ctx, bson.M{"user_id": req.UserId}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.WalletResponse{Balance: 0}, nil
		}
		return nil, fmt.Errorf("mongo find error: %w", err)
	}
	return &pb.WalletResponse{Balance: doc.Balance}, nil
}

// UpdateBalance инкрементирует баланс (upsert)
func (s *walletServer) UpdateBalance(ctx context.Context, req *pb.WalletUpdateRequest) (*pb.WalletUpdateResponse, error) {
	filter := bson.M{"user_id": req.UserId}
	update := bson.M{"$inc": bson.M{"balance": req.Amount}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updated struct {
		Balance int32 `bson:"balance"`
	}
	err := s.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		return nil, fmt.Errorf("mongo update error: %w", err)
	}
	return &pb.WalletUpdateResponse{NewBalance: updated.Balance}, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	ws := NewWalletServer(ctx)
	pb.RegisterWalletServiceServer(s, ws)

	log.Println("Wallet service (MongoDB) running on port 50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
