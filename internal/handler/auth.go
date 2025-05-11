package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/repository"
)

type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

var jwtSecret = getJWTSecret()

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "MY_SUPER_SECRET"
	}
	return []byte(secret)
}

type CustomClaims struct {
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := repository.CreateUser(req.Login, string(hashed)); err != nil {
		http.Error(w, "cannot create user: "+err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, err := repository.GetUserByLogin(req.Login)
	if user == nil || err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	tokenStr, err := generateJWT(user)
	if err != nil {
		http.Error(w, "cannot generate token", http.StatusInternalServerError)
		return
	}

	resp := loginResponse{Token: tokenStr}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func generateJWT(u *model.User) (string, error) {
	claims := &CustomClaims{
		UserID: u.ID,
		Login:  u.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

type contextKey string

const UserIDCtxKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		claims := &CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "token is not valid", http.StatusUnauthorized)
			return
		}

		userID := claims.UserID
		ctx := context.WithValue(r.Context(), UserIDCtxKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(UserIDCtxKey)
	if v == nil {
		return 0, false
	}
	uid, ok := v.(int64)
	return uid, ok
}
