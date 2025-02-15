package routes

import (
	controllers "Handbook/controller"

	"github.com/gin-gonic/gin"
)

func SetupMedicineRoutes(router *gin.Engine) {
	router.GET("/medicines", controllers.GetMedicines)
	router.GET("/medicines/search", controllers.SearchMedicine)
	router.GET("/medicines/category", controllers.GetMedicinesByCategory)
}

func SetupFavoritesRoutes(router *gin.Engine) {
	router.POST("/favorites", controllers.AddToFavorites)
	router.GET("/favorites", controllers.GetFavorites)
	router.DELETE("/favorites/:id", controllers.RemoveFromFavorites)
}

func SetupUserRoutes(router *gin.Engine) {
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.GET("/profile", controllers.GetProfile)
	router.PUT("/update-profile", controllers.UpdateProfile)
	router.DELETE("/delete-user", controllers.DeleteUser)
}

func SetupClinicRoutes(router *gin.Engine) {
	router.GET("/clinics", controllers.GetClinicsByCity)
}
