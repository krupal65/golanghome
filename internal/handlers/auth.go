package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"ghar-kharcha/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", 400)
			return
		}
		if req.Name == "" || req.Email == "" || req.Password == "" {
			jsonError(w, "Sabhi fields bharna zaroori hai", 400)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			jsonError(w, "Server error", 500)
			return
		}

		var userID int
		err = db.QueryRow(
			`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
			req.Name, req.Email, string(hash),
		).Scan(&userID)
		if err != nil {
			jsonError(w, "Email already registered hai", 409)
			return
		}

		// Insert default categories
		defaultCategories := [][]string{
			{"Ghar Ka Kiraya", "🏠", "#ef4444"},
			{"Khaana Peena", "🍽️", "#f97316"},
			{"Bijli Paani", "💡", "#eab308"},
			{"Transport", "🚗", "#3b82f6"},
			{"Kapde", "👕", "#8b5cf6"},
			{"Dawai Ilaj", "💊", "#ec4899"},
			{"Padhai", "📚", "#06b6d4"},
			{"Entertainment", "🎬", "#84cc16"},
			{"EMI Loan", "🏦", "#64748b"},
			{"Other", "📦", "#94a3b8"},
		}
		for _, cat := range defaultCategories {
			db.Exec(`INSERT INTO categories (user_id, name, icon, color) VALUES ($1, $2, $3, $4)`,
				userID, cat[0], cat[1], cat[2])
		}

		token, _ := auth.GenerateToken(userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":   token,
			"user_id": userID,
			"name":    req.Name,
			"email":   req.Email,
		})
	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid request", 400)
			return
		}

		var userID int
		var name, hash string
		err := db.QueryRow(
			`SELECT id, name, password_hash FROM users WHERE email = $1`, req.Email,
		).Scan(&userID, &name, &hash)
		if err != nil {
			jsonError(w, "Email ya password galat hai", 401)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			jsonError(w, "Email ya password galat hai", 401)
			return
		}

		token, _ := auth.GenerateToken(userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":   token,
			"user_id": userID,
			"name":    name,
			"email":   req.Email,
		})
	}
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
