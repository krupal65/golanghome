package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ghar-kharcha/internal/auth"
	"ghar-kharcha/internal/models"

	"github.com/gorilla/mux"
)

func GetIncome(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		month := r.URL.Query().Get("month")
		year := r.URL.Query().Get("year")

		if month == "" || year == "" {
			now := time.Now()
			month = strconv.Itoa(int(now.Month()))
			year = strconv.Itoa(now.Year())
		}

		rows, err := db.Query(
			`SELECT id, amount, description, month, year, created_at FROM income 
			 WHERE user_id=$1 AND month=$2 AND year=$3 ORDER BY created_at DESC`,
			userID, month, year,
		)
		if err != nil {
			jsonError(w, "Server error", 500)
			return
		}
		defer rows.Close()

		var incomes []models.Income
		for rows.Next() {
			var inc models.Income
			inc.UserID = userID
			rows.Scan(&inc.ID, &inc.Amount, &inc.Description, &inc.Month, &inc.Year, &inc.CreatedAt)
			incomes = append(incomes, inc)
		}
		if incomes == nil {
			incomes = []models.Income{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(incomes)
	}
}

func AddIncome(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		var inc models.Income
		if err := json.NewDecoder(r.Body).Decode(&inc); err != nil {
			jsonError(w, "Invalid request", 400)
			return
		}
		if inc.Amount <= 0 {
			jsonError(w, "Amount sahi nahi hai", 400)
			return
		}
		if inc.Month == 0 || inc.Year == 0 {
			now := time.Now()
			inc.Month = int(now.Month())
			inc.Year = now.Year()
		}

		err := db.QueryRow(
			`INSERT INTO income (user_id, amount, description, month, year) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
			userID, inc.Amount, inc.Description, inc.Month, inc.Year,
		).Scan(&inc.ID, &inc.CreatedAt)
		if err != nil {
			jsonError(w, "Server error", 500)
			return
		}
		inc.UserID = userID

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(inc)
	}
}

func DeleteIncome(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		id := mux.Vars(r)["id"]

		res, err := db.Exec(`DELETE FROM income WHERE id=$1 AND user_id=$2`, id, userID)
		if err != nil {
			jsonError(w, "Server error", 500)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			jsonError(w, "Not found", 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}
