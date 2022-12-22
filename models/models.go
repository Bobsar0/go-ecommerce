package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User object
type User struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName      *string            `json:"firstName" validate:"required,min=2,max=30"`
	LastName       *string            `json:"lastName" validate:"required,min=2,max=30"`
	Password       *string            `json:"password" validate:"required,min=6"`
	Email          *string            `json:"email" validate:"required"`
	Phone          *string            `json:"phone" validate:"required"`
	Token          *string            `json:"token"`
	RefreshToken   *string            `json:"refreshToken"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	UserCart       []UserProduct      `json:"userCart" bson:"userCart"`
	AddressDetails []Address          `json:"addressDetails" bson:"addressDetails"`
	OrderStatus    []Order            `json:"orders" bson:"orders"`
	UserId         string             `json:"userId"`
}

type Product struct {
	ProductId   primitive.ObjectID `bson:"_id"`
	ProductName *string            `json:"productName"`
	Price       *uint64            `json:"price"`
	Rating      *uint8             `json:"rating"`
	Image       *string            `json:"image"`
}

type UserProduct struct {
	ProductId   primitive.ObjectID `bson:"_id"`
	ProductName *string            `json:"productName"`
	Price       int                `json:"price"`
	Rating      *uint              `json:"rating"`
	Image       *string            `json:"image"`
}

type Address struct {
	AddressId primitive.ObjectID `bson:"_id"`
	House     *string            `json:"house" bson:"house"`
	Street    *string            `json:"street" bson:"street"`
	City      *string            `json:"city" bson:"city"`
	PinCode   *string            `json:"pinCode" bson:"pinCode"`
}

type Order struct {
	OrderId       primitive.ObjectID `bson:"_id"`
	OrderCart     []UserProduct      `json:"orderList" bson:"orderList"`
	OrderedAt     time.Time          `json:"orderedAt" bson:"orderedAt"`
	Price         int                `json:"totalPrice" bson:"totalPrice"`
	Discount      *int               `json:"discount" bson:"discount"`
	PaymentMethod Payment            `json:"paymentMethod" bson:"paymentMethod"`
}

type Payment struct {
	Digital bool
	COD     bool
}
