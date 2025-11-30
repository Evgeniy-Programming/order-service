package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Order struct {
	OrderUID    string    `json:"order_uid"`
	TrackNumber string    `json:"track_number"`
	Entry       string    `json:"entry"`
	Delivery    Delivery  `json:"delivery"`
	Payment     Payment   `json:"payment"`
	Items       []Item    `json:"items"`
	Locale      string    `json:"locale"`
	CustomerID  string    `json:"customer_id"`
	DateCreated time.Time `json:"date_created"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

// generateRandomOrder создает один случайный, но валидный заказ
func generateRandomOrder() Order {
	trackNumber := gofakeit.Regex("[A-Z]{2}[0-9]{10}[A-Z]{2}")

	var items []Item
	numItems := gofakeit.Number(1, 5)
	goodsTotal := 0
	for i := 0; i < numItems; i++ {
		price := gofakeit.Number(100, 5000)
		sale := gofakeit.Number(5, 50)
		totalPrice := price * (100 - sale) / 100
		goodsTotal += totalPrice
		items = append(items, Item{
			ChrtID:      gofakeit.Number(1000000, 9999999),
			TrackNumber: trackNumber,
			Price:       price,
			Rid:         uuid.New().String(),
			Name:        gofakeit.ProductName(),
			Sale:        sale,
			Size:        "0",
			TotalPrice:  totalPrice,
			NmID:        gofakeit.Number(100000, 999999),
			Brand:       gofakeit.Company(),
			Status:      202,
		})
	}

	deliveryCost := gofakeit.Number(150, 1500)
	orderUID := uuid.New().String()

	return Order{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Address,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: Payment{
			Transaction:  orderUID,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       goodsTotal + deliveryCost,
			PaymentDt:    time.Now().Unix(),
			Bank:         gofakeit.CreditCardType(),
			DeliveryCost: deliveryCost,
			GoodsTotal:   goodsTotal,
		},
		Items:       items,
		Locale:      "en",
		CustomerID:  gofakeit.Username(),
		DateCreated: time.Now(),
	}
}

func main() {
	//парсит аргументы cmd
	count := flag.Int("count", 1, "number of messages to publish")
	flag.Parse()

	if *count <= 0 {
		log.Fatal("Count must be a positive number.")
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP("127.0.0.1:9092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	log.Printf("Starting to publish %d message(s)...", *count)

	//генерация и отправка
	for i := 0; i < *count; i++ {
		order := generateRandomOrder()

		jsonData, err := json.Marshal(order)
		if err != nil {
			log.Printf("ERROR: failed to marshal order: %v", err)
			continue
		}

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Value: jsonData,
			},
		)

		if err != nil {
			log.Printf("ERROR: failed to write message: %v", err)
		} else {
			log.Printf("Successfully published order %s", order.OrderUID)
		}

		//задержка для корр работы
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Finished publishing %d message(s).", *count)
}
