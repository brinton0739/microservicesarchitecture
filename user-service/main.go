package main

import (
	"context"
	"database/sql"
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

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ProfileResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
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

	router := mux.NewRouter()
	// Prefix all routes with /user
	userRouter := router.PathPrefix("/user").Subrouter()
	userRouter.HandleFunc("/register", registerUser).Methods("POST")
	userRouter.HandleFunc("/login", loginUser).Methods("POST")
	userRouter.HandleFunc("/profile", getUserProfile).Methods("GET")
	userRouter.HandleFunc("/users", getUserProfileList).Methods("GET")
	userRouter.HandleFunc("/users/{id}", deleteUserHandler).Methods("DELETE")
	http.ListenAndServe(":8080", router) // Ensure this is set to 8080

	log.Println("User Service is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getDBURL() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	_, err := db.Exec(context.Background(), "INSERT INTO users (username, password, email) VALUES ($1, $2, $3)",
		req.Username, req.Password, req.Email)
	if err != nil {
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered!"))
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow(context.Background(), "SELECT id, username, password, email FROM users WHERE username=$1 AND password=$2",
		req.Username, req.Password).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("User logged in!"))
}

func getUserProfile(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow(context.Background(), "SELECT id, username, email FROM users WHERE username=$1", username).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	resp := ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
	json.NewEncoder(w).Encode(resp)
}

func getUserProfileList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT id, username, email FROM users")
	if err != nil {
		http.Error(w, "Error getting users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []ProfileResponse
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email)
		if err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		resp := ProfileResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
		users = append(users, resp)
	}

	json.NewEncoder(w).Encode(users)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = deleteUser(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "User deleted successfully")

}

func deleteUser(ctx context.Context, userID int) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
