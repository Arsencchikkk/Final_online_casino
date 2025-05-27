// cmd/api_gateway/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gamepb "github.com/Arsencchikkk/casino/proto/game"
	userpb "github.com/Arsencchikkk/casino/proto/user"
	walletpb "github.com/Arsencchikkk/casino/proto/wallet"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	// Подключаемся к gRPC-сервисам
	ua, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial user service:", err)
	}
	ga, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial game service:", err)
	}
	wa, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial wallet service:", err)
	}

	userClient := userpb.NewUserServiceClient(ua)
	gameClient := gamepb.NewGameServiceClient(ga)
	walletClient := walletpb.NewWalletServiceClient(wa)

	// JWT-секрет (в продакшне загружать из os.Getenv)
	secret := []byte("your_super_secret_key_here")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		// === Публичные методы ===

		// Регистрация: создаём пользователя, шлём код, записываем name/surname и сразу даём 1000 кредита
		api.POST("/register", func(c *gin.Context) {
			var body struct {
				Username string `json:"username"`
				Password string `json:"password"`
				Email    string `json:"email"`
				Name     string `json:"name"`
				Surname  string `json:"surname"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// 1) Регистрируем пользователя
			regResp, err := userClient.Register(context.Background(), &userpb.RegisterRequest{
				Username: body.Username,
				Password: body.Password,
				Email:    body.Email,
				Name:     body.Name,
				Surname:  body.Surname,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// 2) Начисляем стартовый баланс 1000
			if _, err := walletClient.UpdateBalance(context.Background(), &walletpb.WalletUpdateRequest{
				UserId: regResp.UserId,
				Amount: 1000,
			}); err != nil {
				log.Printf("warning: cannot set initial balance: %v", err)
			}

			// 3) Отдаём user_id и баланс
			wr, err := walletClient.GetBalance(context.Background(), &walletpb.WalletRequest{
				UserId: regResp.UserId,
			})
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"user_id": regResp.UserId,
					"warning": "user created, but cannot fetch balance",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"user_id": regResp.UserId,
				"balance": wr.Balance,
			})
		})

		// Подтверждение e-mail кодом
		api.POST("/confirm", func(c *gin.Context) {
			var body struct {
				UserId string `json:"user_id"`
				Code   string `json:"code"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			resp, err := userClient.ConfirmEmail(context.Background(), &userpb.ConfirmEmailRequest{
				UserId: body.UserId,
				Code:   body.Code,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !resp.Success {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code or already confirmed"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// Логин: возвращает JWT и user_id
		api.POST("/login", func(c *gin.Context) {
			var body struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			resp, err := userClient.Login(context.Background(), &userpb.LoginRequest{
				Username: body.Username,
				Password: body.Password,
			})
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"token":   resp.Token,
				"user_id": resp.UserId,
			})
		})

		api.GET("/admin/users", func(c *gin.Context) {
			userList, err := userClient.GetAllUsers(context.Background(), &emptypb.Empty{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
				return
			}

			type UserWithBalance struct {
				UserId   string `json:"user_id"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Name     string `json:"name"`
				Surname  string `json:"surname"`
				Balance  int32  `json:"balance"`
			}

			usersWithBalance := make([]UserWithBalance, 0, len(userList.Users))

			for _, u := range userList.Users {
				wr, err := walletClient.GetBalance(context.Background(), &walletpb.WalletRequest{UserId: u.UserId})
				balance := int32(0)
				if err == nil {
					balance = wr.Balance
				}
				usersWithBalance = append(usersWithBalance, UserWithBalance{
					UserId:   u.UserId,
					Username: u.Username,
					Email:    u.Email,
					Name:     u.Name,
					Surname:  u.Surname,
					Balance:  balance,
				})
			}

			c.JSON(http.StatusOK, gin.H{"users": usersWithBalance})
		})

		// === Защищённые методы (JWT) ===
		protected := api.Group("/")
		protected.Use(func(c *gin.Context) {
			auth := c.GetHeader("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
				return
			}
			tokenString := strings.TrimPrefix(auth, "Bearer ")
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			claims := token.Claims.(jwt.MapClaims)
			sub, ok := claims["sub"].(string)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
				return
			}
			c.Set("user_id", sub)
			c.Next()
		})

		// Профиль
		protected.GET("/profile", func(c *gin.Context) {
			uid := c.GetString("user_id")
			resp, err := userClient.GetProfile(context.Background(), &userpb.GetProfileRequest{UserId: uid})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
		})
		protected.PUT("/profile", func(c *gin.Context) {
			var body struct {
				Name     string `json:"name"`
				Surname  string `json:"surname"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			uid := c.GetString("user_id")
			resp, err := userClient.UpdateProfile(context.Background(), &userpb.UpdateProfileRequest{
				UserId:   uid,
				Name:     body.Name,
				Surname:  body.Surname,
				Password: body.Password,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
		})
		protected.DELETE("/profile", func(c *gin.Context) {
			uid := c.GetString("user_id")
			resp, err := userClient.DeleteProfile(context.Background(), &userpb.DeleteProfileRequest{UserId: uid})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
		})

		// Игра + баланс
		protected.POST("/new_game", func(c *gin.Context) {
			uid := c.GetString("user_id")
			gr, err := gameClient.NewGame(context.Background(), &gamepb.NewGameRequest{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			wr, err := walletClient.GetBalance(context.Background(), &walletpb.WalletRequest{UserId: uid})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"session_id":   gr.SessionId,
				"player_cards": gr.PlayerCards,
				"dealer_cards": gr.DealerCards,
				"player_total": gr.PlayerTotal,
				"balance":      wr.Balance,
			})
		})
		protected.POST("/hit", func(c *gin.Context) {
			uid := c.GetString("user_id")
			sid := c.Query("session_id")
			hr, err := gameClient.Hit(context.Background(), &gamepb.HitRequest{SessionId: sid})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			wr, err := walletClient.GetBalance(context.Background(), &walletpb.WalletRequest{UserId: uid})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"player_cards": hr.PlayerCards,
				"player_total": hr.PlayerTotal,
				"finished":     hr.Finished,
				"balance":      wr.Balance,
			})
		})
		protected.POST("/stand", func(c *gin.Context) {
			uid := c.GetString("user_id")
			sid := c.Query("session_id")
			sr, err := gameClient.Stand(context.Background(), &gamepb.StandRequest{SessionId: sid})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// корректировка баланса
			var delta int32
			switch sr.Outcome {
			case "win":
				delta = 200
			case "lose":
				delta = -100
			}
			wur, err := walletClient.UpdateBalance(context.Background(), &walletpb.WalletUpdateRequest{
				UserId: uid,
				Amount: delta,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"dealer_cards": sr.DealerCards,
				"dealer_total": sr.DealerTotal,
				"outcome":      sr.Outcome,
				"balance":      wur.NewBalance,
			})
		})
		protected.GET("/wallet", func(c *gin.Context) {
			uid := c.GetString("user_id")
			wr, err := walletClient.GetBalance(context.Background(), &walletpb.WalletRequest{UserId: uid})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"balance": wr.Balance})
		})

	}

	admin := r.Group("/api/admin")
	{
		admin.PUT("/users/:id", func(c *gin.Context) {
			userID := c.Param("id")
			var body struct {
				Name     string `json:"name"`
				Surname  string `json:"surname"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			resp, err := userClient.UpdateProfile(context.Background(), &userpb.UpdateProfileRequest{
				UserId:   userID,
				Name:     body.Name,
				Surname:  body.Surname,
				Password: body.Password,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
		})

		admin.DELETE("/users/:id", func(c *gin.Context) {
			userID := c.Param("id")
			resp, err := userClient.DeleteProfile(context.Background(), &userpb.DeleteProfileRequest{
				UserId: userID,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
		})
	}

	// Отдаём SPA
	r.GET("/", func(c *gin.Context) {
		c.File("../Front/index.html")
	})

	r.GET("/admin.html", func(c *gin.Context) {
		c.File("../Front/admin.html")
	})

	// В main.go или router.go
	r.PUT("/api/admin/users/:id/balance", func(c *gin.Context) {
		userID := c.Param("id")

		var body struct {
			Amount int32 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Вызов gRPC WalletService.UpdateBalance
		resp, err := walletClient.UpdateBalance(context.Background(), &walletpb.WalletUpdateRequest{
			UserId: userID,
			Amount: body.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":     userID,
			"new_balance": resp.NewBalance,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("API-Gateway listening on :%s …", port)
	log.Fatal(srv.ListenAndServe())
}
