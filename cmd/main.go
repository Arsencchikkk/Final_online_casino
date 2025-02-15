package main

import (
	"Handbook/config"
	"Handbook/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	config.ConnectDatabase()
	config.AutoMigrate()

	routes.SetupUserRoutes(router)
	routes.SetupMedicineRoutes(router)
	routes.SetupFavoritesRoutes(router)
	routes.SetupClinicRoutes(router)

	router.Static("/static", "./Front")

	router.GET("/", func(c *gin.Context) {
		c.File("./Front/index.html")
	})

	router.GET("/login.html", func(c *gin.Context) {
		c.File("./Front/login.html")
	})

	router.Run(":8080")
}
