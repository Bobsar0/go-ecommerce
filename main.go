package main

import (
	"go-ecommerce/routes"
	"os"

	db "go-ecommerce/database"
"go-ecommerce/middleware"
"go-ecommerce/controllers"
	"github.com/gin-gonic/gin"
)

func main()  {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := controllers.NewApplication(db.ProductData(db.Client, "Products"), db.UserData(db.Client, ""))

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	router.GET("/addtocart", app.AddToCart())
		router.GET("/removeitem", app.RemoveItem())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

}