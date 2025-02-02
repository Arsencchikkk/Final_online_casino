package main

import (
	"Handbook/config" // убедись, что модуль написан строчными буквами
	"Handbook/middleware"
	"Handbook/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Добавляем CORS Middleware
	router.Use(middleware.CORSMiddleware())

	// Подключаем базу данных
	config.ConnectDatabase()

	// Настраиваем маршруты API
	routes.SetupMedicineRoutes(router)

	// Раздаём статические файлы (Frontend)
	router.Static("/static", "./Front")

	// Главная страница (возвращает index.html)
	router.GET("/", func(c *gin.Context) {
		c.File("./Front/index.html")
	})

	// Запускаем сервер на порту 8080
	router.Run(":8080")
}
