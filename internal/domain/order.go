package domain

import "time"

type Order struct {
	OrderUID    string    `json:"order_uid" validate:"required"`
	TrackNumber string    `json:"track_number" validate:"required"`
	Entry       string    `json:"entry" validate:"required"`
	Delivery    Delivery  `json:"delivery" validate:"required"`
	Payment     Payment   `json:"payment" validate:"required"`
	Items       []Item    `json:"items" validate:"required,min=1"`
	Locale      string    `json:"locale" validate:"required"`
	CustomerID  string    `json:"customer_id" validate:"required"`
	DateCreated time.Time `json:"date_created" validate:"required"`
}

type Delivery struct {
	Name    string `json:"name" validate:"required"`
	Phone   string `json:"phone" validate:"required,e164"` // для международних номеров
	Zip     string `json:"zip" validate:"required"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency" validate:"required,iso4217"` //iso4217 - код валюты (USD, RUB, etc.)
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"gte=0"` //сумма не отриц
	PaymentDt    int64  `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" validate:"gte=0"`
	CustomFee    int    `json:"custom_fee" validate:"gte=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"gte=0"`
	Rid         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"gte=0,lte=100"` //скидка от 0 до 100
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price" validate:"gte=0"`
	NmID        int    `json:"nm_id" validate:"required"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required"`
}
