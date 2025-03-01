package routes

import(
	"go-ecommerce/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/userssignup", controllers.SignUp())
	incomingRoutes.POST("/userslogin", controllers.Login())
	//incomingRoutes.POST("/users/addproduct", controllers.ProductViewerAdmin())
	incomingRoutes.GET("/users/productview", controllers.SearchProduct())
	incomingRoutes.GET("/users/search", controllers.SearchProductByQuery())
}