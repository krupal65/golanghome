package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ghar-kharcha/internal/auth"
	"ghar-kharcha/internal/models"
)

func GetSummary(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		monthStr := r.URL.Query().Get("month")
		yearStr := r.URL.Query().Get("year")

		now := time.Now()
		month := int(now.Month())
		year := now.Year()

		if monthStr != "" {
			month, _ = strconv.Atoi(monthStr)
		}
		if yearStr != "" {
			year, _ = strconv.Atoi(yearStr)
		}

		summary := models.Summary{Month: month, Year: year}

		// Total income
		db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM income WHERE user_id=$1 AND month=$2 AND year=$3`,
			userID, month, year).Scan(&summary.TotalIncome)

		// Total expenses
		db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM expenses WHERE user_id=$1 AND month=$2 AND year=$3`,
			userID, month, year).Scan(&summary.TotalExpenses)

		summary.Balance = summary.TotalIncome - summary.TotalExpenses

		// Category breakdown
		rows, err := db.Query(
			`SELECT COALESCE(c.name,'Other'), COALESCE(c.icon,'📦'), COALESCE(c.color,'#94a3b8'), 
			        COALESCE(SUM(e.amount),0) as total
			 FROM expenses e
			 LEFT JOIN categories c ON c.id = e.category_id
			 WHERE e.user_id=$1 AND e.month=$2 AND e.year=$3
			 GROUP BY c.name, c.icon, c.color
			 ORDER BY total DESC`,
			userID, month, year,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var cs models.CategoryStat
				rows.Scan(&cs.CategoryName, &cs.CategoryIcon, &cs.CategoryColor, &cs.Total)
				if summary.TotalExpenses > 0 {
					cs.Percentage = (cs.Total / summary.TotalExpenses) * 100
				}
				summary.CategoryBreakdown = append(summary.CategoryBreakdown, cs)
			}
		}
		if summary.CategoryBreakdown == nil {
			summary.CategoryBreakdown = []models.CategoryStat{}
		}

		// Monthly trend - last 6 months
		trendRows, err := db.Query(
			`SELECT month, year, 
			        COALESCE((SELECT SUM(amount) FROM income i WHERE i.user_id=$1 AND i.month=m.month AND i.year=m.year),0),
			        COALESCE((SELECT SUM(amount) FROM expenses ex WHERE ex.user_id=$1 AND ex.month=m.month AND ex.year=m.year),0)
			 FROM (
			   SELECT generate_series(1,12) as month, $2 as year
			   UNION ALL
			   SELECT generate_series(1,12), $2-1
			 ) m
			 WHERE (m.year < $2) OR (m.year = $2 AND m.month <= $3)
			 ORDER BY m.year DESC, m.month DESC
			 LIMIT 6`,
			userID, year, month,
		)
		if err == nil {
			defer trendRows.Close()
			for trendRows.Next() {
				var ms models.MonthStat
				trendRows.Scan(&ms.Month, &ms.Year, &ms.Income, &ms.Expense)
				summary.MonthlyTrend = append(summary.MonthlyTrend, ms)
			}
			// Reverse for chronological order
			for i, j := 0, len(summary.MonthlyTrend)-1; i < j; i, j = i+1, j-1 {
				summary.MonthlyTrend[i], summary.MonthlyTrend[j] = summary.MonthlyTrend[j], summary.MonthlyTrend[i]
			}
		}
		if summary.MonthlyTrend == nil {
			summary.MonthlyTrend = []models.MonthStat{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}

func GetCategories(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		rows, err := db.Query(`SELECT id, name, icon, color FROM categories WHERE user_id=$1 ORDER BY name`, userID)
		if err != nil {
			jsonError(w, "Server error", 500)
			return
		}
		defer rows.Close()

		var cats []models.Category
		for rows.Next() {
			var c models.Category
			c.UserID = userID
			rows.Scan(&c.ID, &c.Name, &c.Icon, &c.Color)
			cats = append(cats, c)
		}
		if cats == nil {
			cats = []models.Category{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cats)
	}
}

func AddCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r)
		var cat models.Category
		if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
			jsonError(w, "Invalid request", 400)
			return
		}
		if cat.Name == "" {
			jsonError(w, "Category name required", 400)
			return
		}
		if cat.Icon == "" {
			cat.Icon = "📦"
		}
		if cat.Color == "" {
			cat.Color = "#94a3b8"
		}

		err := db.QueryRow(
			`INSERT INTO categories (user_id, name, icon, color) VALUES ($1,$2,$3,$4) RETURNING id`,
			userID, cat.Name, cat.Icon, cat.Color,
		).Scan(&cat.ID)
		if err != nil {
			jsonError(w, "Server error", 500)
			return
		}
		cat.UserID = userID

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(cat)
	}
}
