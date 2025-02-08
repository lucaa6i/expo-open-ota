package middleware

import (
	"expo-open-ota/internal/auth"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No Authorization header provided", http.StatusUnauthorized)
			return
		}
		bearerToken := authHeader[len("Bearer "):]
		authService := auth.NewAuth()
		_, err := authService.ValidateToken(bearerToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
