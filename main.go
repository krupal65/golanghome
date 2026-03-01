package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ghar-kharcha/internal/auth"
	"ghar-kharcha/internal/database"
	"ghar-kharcha/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	db, err := database.Connect()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatal("Migration failed:", err)
	}

	r := mux.NewRouter()
	authMiddleware := auth.Middleware

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/static"))))
	r.HandleFunc("/", serveIndex).Methods("GET")

	r.HandleFunc("/api/register", handlers.Register(db)).Methods("POST")
	r.HandleFunc("/api/login", handlers.Login(db)).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware)

	api.HandleFunc("/income", handlers.GetIncome(db)).Methods("GET")
	api.HandleFunc("/income", handlers.AddIncome(db)).Methods("POST")
	api.HandleFunc("/income/{id}", handlers.DeleteIncome(db)).Methods("DELETE")

	api.HandleFunc("/expenses", handlers.GetExpenses(db)).Methods("GET")
	api.HandleFunc("/expenses", handlers.AddExpense(db)).Methods("POST")
	api.HandleFunc("/expenses/{id}", handlers.UpdateExpense(db)).Methods("PUT")
	api.HandleFunc("/expenses/{id}", handlers.DeleteExpense(db)).Methods("DELETE")

	api.HandleFunc("/summary", handlers.GetSummary(db)).Methods("GET")
	api.HandleFunc("/categories", handlers.GetCategories(db)).Methods("GET")
	api.HandleFunc("/categories", handlers.AddCategory(db)).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Ghar Kharcha server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/index.html")
}
