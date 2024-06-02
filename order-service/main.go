package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

var (
	db *pgx.Conn
	mu sync.Mutex
)

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"`
	Total     float64   `json:"total"`
	OrderDate time.Time `json:"order_date"`
}

func main() {
	var err error
	maxRetries := 10
	dbURL := getDBURL()
	for i := 0; i < maxRetries; i++ {
		db, err = pgx.Connect(context.Background(), dbURL)
		if err == nil {
			break
		}
		log.Printf("Unable to connect to database (attempt %d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Unable to connect to database after %d attempts: %v\n", maxRetries, err)
	}
	defer db.Close(context.Background())

	r := mux.NewRouter()
	subrouter := r.PathPrefix("/order").Subrouter()
	subrouter.HandleFunc("/order", createOrder).Methods("POST")
	subrouter.HandleFunc("/orders", getOrders).Methods("GET")
	subrouter.HandleFunc("/order/{id}", getOrderByID).Methods("GET")

	log.Println("Order Service is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getDBURL() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	_, err := db.Exec(context.Background(), "INSERT INTO orders (user_id, product_id, quantity, status, total) VALUES ($1, $2, $3, $4, $5)",
		order.UserID, order.ProductID, order.Quantity, order.Status, order.Total)
	if err != nil {
		http.Error(w, "Error creating order", http.StatusInternalServerError)
		fmt.Println(err) // HERE - print the error to the console
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Order created!"))
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT id, user_id, product_id, quantity, status, total, order_date FROM orders")
	if err != nil {
		http.Error(w, "Error fetching orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.ProductID, &order.Quantity, &order.Status, &order.Total, &order.OrderDate); err != nil {
			http.Error(w, "Error scanning order", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		orders = append(orders, order)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func getOrderByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var order Order
	err = db.QueryRow(context.Background(), "SELECT id, user_id, product_id, quantity, status, total, order_date FROM orders WHERE id=$1", id).Scan(&order.ID, &order.UserID, &order.ProductID, &order.Quantity, &order.Status, &order.OrderDate)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
