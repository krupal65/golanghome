package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

type Income struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Month       int       `json:"month"`
	Year        int       `json:"year"`
	CreatedAt   time.Time `json:"created_at"`
}

type Expense struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	CategoryID   *int      `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	CategoryIcon string    `json:"category_icon,omitempty"`
	CategoryColor string   `json:"category_color,omitempty"`
	Amount       float64   `json:"amount"`
	Description  string    `json:"description"`
	Date         string    `json:"date"`
	Month        int       `json:"month"`
	Year         int       `json:"year"`
	CreatedAt    time.Time `json:"created_at"`
}

type Summary struct {
	Month          int                `json:"month"`
	Year           int                `json:"year"`
	TotalIncome    float64            `json:"total_income"`
	TotalExpenses  float64            `json:"total_expenses"`
	Balance        float64            `json:"balance"`
	CategoryBreakdown []CategoryStat  `json:"category_breakdown"`
	MonthlyTrend   []MonthStat        `json:"monthly_trend"`
}

type CategoryStat struct {
	CategoryName  string  `json:"category_name"`
	CategoryIcon  string  `json:"category_icon"`
	CategoryColor string  `json:"category_color"`
	Total         float64 `json:"total"`
	Percentage    float64 `json:"percentage"`
}

type MonthStat struct {
	Month   int     `json:"month"`
	Year    int     `json:"year"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}
