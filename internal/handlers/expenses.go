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

func GetExpenses(db *sql.DB) http.HandlerFunc {
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
			`SELECT e.id, e.amount, e.description, e.date, e.month, e.year, e.created_at,
			        e.category_id, COALESCE(c.name,'Other'), COALESCE(c.icon,'📦'), COALESCE(c.color,'#94a3b8')
			 FROM expenses e
			 LEFT JOIN categories c ON c.id = e.category_id
			 WHERE e.user_id=$1 AND e.month=$2 AND e.year=$3
			 ORDER BY e.date DESC, e.created_at DESC`,
			userID, month, year,
		)
		if err != nil {
			jsonError(w, "Server error", 500)
			return
		}
		defer rows.Close()

		var expenses []models.Expense
		for rows.Next() {
			var exp models.Expense
			exp.UserID = userID
			rows.Scan(&exp.ID, &exp.Amount, &exp.Description, &exp.Date, &exp.Month, &exp.Year, &exp.CreatedAt,
				&exp.CategoryID, &exp.CategoryName, &exp.CategoryIcon, &exp.CategoryColor)
			expenses = append(expenses, exp)
		}
		if expenses == nil {
			expenses = []models.Expense{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expenses)
	}
}

func AddExpense(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		var exp models.Expense
		if err := json.NewDecoder(r.Body).Decode(&exp); err != nil {
			jsonError(w, "Invalid request", 400)
			return
		}
		if exp.Amount <= 0 {
			jsonError(w, "Amount sahi nahi hai", 400)
			return
		}
		if exp.Date == "" {
			exp.Date = time.Now().Format("2006-01-02")
		}
		// Parse month/year from date
		t, err := time.Parse("2006-01-02", exp.Date)
		if err != nil {
			jsonError(w, "Date format galat hai (YYYY-MM-DD)", 400)
			return
		}
		exp.Month = int(t.Month())
		exp.Year = t.Year()

		err = db.QueryRow(
			`INSERT INTO expenses (user_id, category_id, amount, description, date, month, year) 
			 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
			userID, exp.CategoryID, exp.Amount, exp.Description, exp.Date, exp.Month, exp.Year,
		).Scan(&exp.ID, &exp.CreatedAt)
		if err != nil {
			jsonError(w, "Server error: "+err.Error(), 500)
			return
		}
		exp.UserID = userID

		// Fetch category details
		if exp.CategoryID != nil {
			db.QueryRow(`SELECT name, icon, color FROM categories WHERE id=$1`, exp.CategoryID).
				Scan(&exp.CategoryName, &exp.CategoryIcon, &exp.CategoryColor)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(exp)
	}
}

func UpdateExpense(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		id := mux.Vars(r)["id"]

		var exp models.Expense
		if err := json.NewDecoder(r.Body).Decode(&exp); err != nil {
			jsonError(w, "Invalid request", 400)
			return
		}

		t, err := time.Parse("2006-01-02", exp.Date)
		if err != nil {
			jsonError(w, "Date format galat hai", 400)
			return
		}
		exp.Month = int(t.Month())
		exp.Year = t.Year()

		res, err := db.Exec(
			`UPDATE expenses SET category_id=$1, amount=$2, description=$3, date=$4, month=$5, year=$6
			 WHERE id=$7 AND user_id=$8`,
			exp.CategoryID, exp.Amount, exp.Description, exp.Date, exp.Month, exp.Year, id, userID,
		)
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

func DeleteExpense(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		id := mux.Vars(r)["id"]

		res, err := db.Exec(`DELETE FROM expenses WHERE id=$1 AND user_id=$2`, id, userID)
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
