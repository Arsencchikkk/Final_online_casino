package main

import (
	"context"
	"testing"

	userpb "github.com/Arsencchikkk/casino/proto/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мокаем интерфейс UserServiceServer
type MockUserService struct {
	mock.Mock
	userpb.UnimplementedUserServiceServer
}

func (m *MockUserService) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*userpb.RegisterResponse), args.Error(1)
}

func TestRegister(t *testing.T) {
	mockService := new(MockUserService)

	req := &userpb.RegisterRequest{
		Username: "testuser",
		Password: "secure123",
		Email:    "test@example.com",
		Name:     "John",
		Surname:  "Doe",
	}

	expectedResp := &userpb.RegisterResponse{
		UserId: "123456",
	}

	// Указываем ожидаемый вызов
	mockService.On("Register", mock.Anything, req).Return(expectedResp, nil)

	// Вызываем Register
	res, err := mockService.Register(context.Background(), req)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, expectedResp.UserId, res.UserId)

	// Проверяем, был ли вызов
	mockService.AssertCalled(t, "Register", mock.Anything, req)
}
