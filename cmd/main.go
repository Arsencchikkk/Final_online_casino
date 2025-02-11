package main

import (
	"Handbook/config"
	"Handbook/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Разрешаем CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	config.ConnectDatabase()
	config.AutoMigrate()

	// Настройка маршрутов
	routes.SetupUserRoutes(router)
	routes.SetupMedicineRoutes(router)
	routes.SetupFavoritesRoutes(router)

	// Раздача статических файлов
	router.Static("/static", "./Front")

	// Раздаём главную страницу (index.html)
	router.GET("/", func(c *gin.Context) {
		c.File("./Front/index.html")
	})

	// Раздаём login.html
	router.GET("/login.html", func(c *gin.Context) {
		c.File("./Front/login.html")
	})

	router.Run(":8080")
}
