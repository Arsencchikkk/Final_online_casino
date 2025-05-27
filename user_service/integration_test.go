package main

import (
	"context"
	"log"
	"testing"
	"time"

	userpb "github.com/Arsencchikkk/casino/proto/user"
	"google.golang.org/grpc"
)

func TestIntegrationRegister(t *testing.T) {
	// Устанавливаем соединение с реальным gRPC-сервером
	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("не удалось подключиться к серверу: %v", err)
	}
	defer conn.Close()

	client := userpb.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &userpb.RegisterRequest{
		Username: "integration_test_user",
		Password: "test123",
		Email:    "test_integration@example.com",
		Name:     "Test",
		Surname:  "Integration",
	}

	res, err := client.Register(ctx, req)
	if err != nil {
		t.Fatalf("ошибка при регистрации: %v", err)
	}

	if res.GetUserId() == "" {
		t.Errorf("ожидался user_id, но получен пустой: %v", res)
	} else {
		log.Printf("✅ Успешно зарегистрирован user_id: %s", res.GetUserId())
	}
}
