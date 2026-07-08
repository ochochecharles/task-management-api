package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
	"database/sql"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ochochecharles/task-management-api/internal/db"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	queries     *db.Queries
	oauthConfig *oauth2.Config
}

func NewAuthHandler(queries *db.Queries) *AuthHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthHandler{
		queries:     queries,
		oauthConfig: oauthConfig,
	}
}

// GoogleLogin redirects the user to Google's OAuth consent page
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the redirect back from Google
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// exchange the code for a Google token
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("failed to exchange token", "error", err)
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	// fetch the user's Google profile
	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		slog.Error("failed to fetch user info", "error", err)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		slog.Error("failed to decode user info", "error", err)
		http.Error(w, "failed to decode user info", http.StatusInternalServerError)
		return
	}

	// find or create the user in our database
	user, err := h.queries.GetUserByGoogleID(r.Context(), googleUser.ID)
	if err != nil {
		// user doesn't exist yet, create them
		user, err = h.queries.CreateUser(r.Context(), db.CreateUserParams{
			GoogleID:  googleUser.ID,
			Email:     googleUser.Email,
			Name:      googleUser.Name,
			AvatarUrl: sql.NullString{
				String: googleUser.AvatarURL,
				Valid:  googleUser.AvatarURL != "",
			},
		})
		if err != nil {
			slog.Error("failed to create user", "error", err)
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		slog.Error("failed to sign token", "error", err)
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	w.Header().Set("Content-Type", "application/json")
	http.Redirect(w, r, frontendURL+"/auth/callback?token="+tokenString, http.StatusTemporaryRedirect)
	// json.NewEncoder(w).Encode(map[string]string{
	// 	"token": tokenString,
	// })
}