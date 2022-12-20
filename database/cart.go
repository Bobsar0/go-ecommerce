package database

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrCantFindProduct = errors.New("cant find product")
	ErrCantDecodeProduct = errors.New("cant decode product")
	ErrUserIdIsNotValid = errors.New("user id is not valid")
	ErrCantUpdateUser = errors.New("cant update user")
	ErrCantRemoveItemCart = errors.New("cant remove item from cart")
	ErrCantGetItem = errors.New("cant get item")
	ErrCantBuyCartItem = errors.New("cant buy cart item")
)

func AddToCart() gin.HandlerFunc {
	
}

func RemoveItem() gin.HandlerFunc {
	
}

func GetItemFromCat() gin.HandlerFunc {
	
}
func BuyFromCart()  gin.HandlerFunc{
	
}

func InstantBuy() gin.HandlerFunc {
	
}