package routes

import (
	controllers "Handbook/controller"

	"github.com/gin-gonic/gin"
)

func SetupMedicineRoutes(router *gin.Engine) {
	router.GET("/medicines", controllers.GetMedicines)
	router.GET("/medicines/search", controllers.SearchMedicine)
}
