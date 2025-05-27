package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"time"

	userpb "github.com/Arsencchikkk/casino/proto/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userServer struct {
	userpb.UnimplementedUserServiceServer
	col       *mongo.Collection
	jwtKey    []byte
	smtpHost  string
	smtpPort  string
	smtpUser  string
	smtpPass  string
	emailFrom string
}

func NewUserServer(ctx context.Context) *userServer {
	_ = godotenv.Load()

	uri := os.Getenv("MONGO_URI")
	db := os.Getenv("MONGO_DB")
	uc := os.Getenv("MONGO_USER_COL")
	secret := os.Getenv("JWT_SECRET")
	if uri == "" || db == "" || uc == "" || secret == "" {
		log.Fatal("env MONGO_URI, MONGO_DB, MONGO_USER_COL, JWT_SECRET required")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	emailFrom := os.Getenv("EMAIL_FROM")
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || emailFrom == "" {
		log.Fatal("env SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, EMAIL_FROM required")
	}

	return &userServer{
		col:       client.Database(db).Collection(uc),
		jwtKey:    []byte(secret),
		smtpHost:  smtpHost,
		smtpPort:  smtpPort,
		smtpUser:  smtpUser,
		smtpPass:  smtpPass,
		emailFrom: emailFrom,
	}
}

// genCode возвращает случайный 6-значный hex-код
func genCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// sendEmail шлёт письмо с кодом подтверждения
func (s *userServer) sendEmail(to, code string) error {
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Ваш код подтверждения\r\n"+
			"\r\n"+
			"Код: %s\r\n",
		s.emailFrom, to, code,
	))
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	// envelope sender — указываем emailFrom
	return smtp.SendMail(addr, auth, s.emailFrom, []string{to}, msg)
}

func (s *userServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	// 1) Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// 2) Генерируем код и сохраняем пользователя с verified=false
	code := genCode()
	res, err := s.col.InsertOne(ctx, bson.M{
		"username":      req.Username,
		"password_hash": string(hash),
		"email":         req.Email,   // ←
		"name":          req.Name,    // ←
		"surname":       req.Surname, // ←
		"verified":      false,
		"code":          code,
	})
	if err != nil {
		return nil, err
	}
	uid := res.InsertedID.(primitive.ObjectID).Hex()

	// 3) Отправляем письмо
	if err := s.sendEmail(req.Email, code); err != nil {
		log.Printf("warn: email send failed: %v", err)
	}
	return &userpb.RegisterResponse{UserId: uid}, nil
}

func (s *userServer) ConfirmEmail(ctx context.Context, req *userpb.ConfirmEmailRequest) (*userpb.ConfirmEmailResponse, error) {
	oid, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": oid, "code": req.Code}
	// При unset лучше использовать любое ненулевое значение, например 1
	update := bson.M{
		"$set":   bson.M{"verified": true},
		"$unset": bson.M{"code": 1},
	}
	res, err := s.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return &userpb.ConfirmEmailResponse{
		Success: res.ModifiedCount == 1,
	}, nil
}

func (s *userServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	var doc struct {
		ID           primitive.ObjectID `bson:"_id"`
		Username     string             `bson:"username"`
		PasswordHash string             `bson:"password_hash"`
		Verified     bool               `bson:"verified"`
	}
	err := s.col.FindOne(ctx, bson.M{"username": req.Username}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("invalid credentials")
	} else if err != nil {
		return nil, err
	}
	if !doc.Verified {
		return nil, fmt.Errorf("email not verified")
	}
	if bcrypt.CompareHashAndPassword([]byte(doc.PasswordHash), []byte(req.Password)) != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	claims := jwt.MapClaims{
		"sub": doc.ID.Hex(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(s.jwtKey)
	if err != nil {
		return nil, err
	}
	return &userpb.LoginResponse{Token: ss, UserId: doc.ID.Hex()}, nil
}

func (s *userServer) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
	oid, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Username string `bson:"username"`
		Email    string `bson:"email"`
		Name     string `bson:"name"`
		Surname  string `bson:"surname"`
	}
	if err := s.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		return nil, err
	}
	return &userpb.GetProfileResponse{
		UserId:   req.UserId,
		Username: doc.Username,
		Email:    doc.Email,
		Name:     doc.Name,
		Surname:  doc.Surname,
	}, nil
}

func (s *userServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	oid, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, err
	}
	set := bson.M{"name": req.Name, "surname": req.Surname}
	if req.Password != "" {
		h, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		set["password_hash"] = string(h)
	}
	update := bson.M{"$set": set}
	res, err := s.col.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return nil, err
	}
	return &userpb.UpdateProfileResponse{Success: res.ModifiedCount == 1}, nil
}

func (s *userServer) DeleteProfile(ctx context.Context, req *userpb.DeleteProfileRequest) (*userpb.DeleteProfileResponse, error) {
	oid, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, err
	}
	res, err := s.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, err
	}
	return &userpb.DeleteProfileResponse{Success: res.DeletedCount == 1}, nil
}

func (s *userServer) GetAllUsers(ctx context.Context, _ *emptypb.Empty) (*userpb.UserList, error) {
	cursor, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []*userpb.User

	for cursor.Next(ctx) {
		var doc struct {
			ID       primitive.ObjectID `bson:"_id"`
			Username string             `bson:"username"`
			Email    string             `bson:"email"`
			Name     string             `bson:"name"`
			Surname  string             `bson:"surname"`
		}

		if err := cursor.Decode(&doc); err != nil {
			continue // или log.Println(err)
		}

		users = append(users, &userpb.User{
			UserId:   doc.ID.Hex(),
			Username: doc.Username,
			Email:    doc.Email,
			Name:     doc.Name,
			Surname:  doc.Surname,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return &userpb.UserList{Users: users}, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatal(err)
	}
	grpcSrv := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcSrv, NewUserServer(ctx))
	log.Println("UserService running on :50053")
	log.Fatal(grpcSrv.Serve(lis))
}
