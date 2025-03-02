package middleware

import(
	"log"
	"strings"
	token "go-ecommerce/tokens"
	"net/http"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc{
	return func(c *gin.Context){
		authHeader := c.Request.Header.Get("Authorization")
		log.Println("Authorization Header:", authHeader)

		if authHeader == ""{
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "No Authorization Header",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization Format",
			})
			c.Abort()
			return
		}

		// Extract the actual token
		clientToken := parts[1]


		claims, err := token.ValidateToken(clientToken)
		if err != ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("uid", claims.Uid)
		c.Next()
	}
}