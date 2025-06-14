package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	walletpb "github.com/Arsencchikkk/final/casino/proto/wallet"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

// WalletDoc представляет документ в Mongo
type WalletDoc struct {
	UserId  string `bson:"user_id"`
	Balance int32  `bson:"balance"`
}

type server struct {
	walletpb.UnimplementedWalletServiceServer
	mongoCol *mongo.Collection
	redis    *redis.Client
}

func NewServer(ctx context.Context) *server {
	// грузим .env (если он есть в текущей папке)
	_ = godotenv.Load()

	// читаем переменные
	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	mongoCol := os.Getenv("MONGO_COLLECTION")
	redisURL := os.Getenv("REDIS_URL")

	if mongoURI == "" || mongoDB == "" || mongoCol == "" || redisURL == "" {
		log.Fatal("MONGO_URI, MONGO_DB, MONGO_COLLECTION и REDIS_URL должны быть заданы")
	}

	// MongoDB
	log.Printf("[init] connecting to MongoDB at %s", mongoURI)
	mClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("[init][mongo] connect error: %v", err)
	}
	col := mClient.Database(mongoDB).Collection(mongoCol)
	log.Printf("[init] MongoDB connected: DB=%s, COLLECTION=%s", mongoDB, mongoCol)

	// Redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("[init][redis] parse URL error: %v", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("[init][redis] ping error: %v", err)
	}
	log.Printf("[init] Redis connected at %s", redisURL)

	return &server{mongoCol: col, redis: rdb}
}

func (s *server) GetBalance(ctx context.Context, req *walletpb.WalletRequest) (*walletpb.WalletResponse, error) {
	key := "balance:" + req.UserId
	log.Printf("[GetBalance] user=%s", req.UserId)

	// 1) пробуем кеш
	if val, err := s.redis.Get(ctx, key).Result(); err == nil {
		log.Printf("[GetBalance] cache hit: %s=%s", key, val)
		if b, err := strconv.Atoi(val); err == nil {
			return &walletpb.WalletResponse{Balance: int32(b)}, nil
		}
	} else if err != redis.Nil {
		log.Printf("[GetBalance] redis GET error: %v", err)
	}

	// 2) кеш-промах — читаем из Mongo
	log.Printf("[GetBalance] cache miss, query MongoDB user=%s", req.UserId)
	filter := bson.M{"user_id": req.UserId}
	var doc WalletDoc
	err := s.mongoCol.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		log.Printf("[GetBalance] no wallet, create default for %s", req.UserId)
		doc = WalletDoc{UserId: req.UserId, Balance: 0}
		if _, err := s.mongoCol.InsertOne(ctx, doc); err != nil {
			log.Printf("[GetBalance] insert default error: %v", err)
			return nil, err
		}
	} else if err != nil {
		log.Printf("[GetBalance] mongo FIND error: %v", err)
		return nil, err
	}

	// 3) записываем в кеш
	log.Printf("[GetBalance] caching %s=%d", key, doc.Balance)
	if err := s.redis.Set(ctx, key, doc.Balance, 5*time.Minute).Err(); err != nil {
		log.Printf("[GetBalance] redis SET error: %v", err)
	}

	return &walletpb.WalletResponse{Balance: doc.Balance}, nil
}

func (s *server) UpdateBalance(ctx context.Context, req *walletpb.WalletUpdateRequest) (*walletpb.WalletUpdateResponse, error) {
	key := "balance:" + req.UserId
	log.Printf("[UpdateBalance] user=%s delta=%d", req.UserId, req.Amount)

	// атомарное обновление в Mongo
	filter := bson.M{"user_id": req.UserId}
	update := bson.M{"$inc": bson.M{"balance": req.Amount}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var updated WalletDoc
	if err := s.mongoCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated); err != nil {
		log.Printf("[UpdateBalance] mongo FindOneAndUpdate error: %v", err)
		return nil, err
	}
	log.Printf("[UpdateBalance] new Mongo balance for %s = %d", req.UserId, updated.Balance)

	// обновляем кеш
	log.Printf("[UpdateBalance] setting cache %s=%d", key, updated.Balance)
	if err := s.redis.Set(ctx, key, updated.Balance, 5*time.Minute).Err(); err != nil {
		log.Printf("[UpdateBalance] redis SET error: %v", err)
	}

	return &walletpb.WalletUpdateResponse{NewBalance: updated.Balance}, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv := NewServer(ctx)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	walletpb.RegisterWalletServiceServer(grpcSrv, srv)

	log.Println("WalletService running on :50052")
	log.Fatal(grpcSrv.Serve(lis))
}
