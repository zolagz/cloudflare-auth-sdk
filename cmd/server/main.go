package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	
	"github.com/cloudflare/cloudflare-go"
	
	"github.com/zolagz/cloudflare-auth-sdk/internal/auth"
	"github.com/zolagz/cloudflare-auth-sdk/internal/config"
	"github.com/zolagz/cloudflare-auth-sdk/internal/kv"
	apperrors "github.com/zolagz/cloudflare-auth-sdk/internal/errors"
)

type Server struct {
	authService *auth.Service
	kvClient    *kv.Client
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Initialize Cloudflare API client
	var api *cloudflare.API
	if cfg.CloudflareAPIToken != "" {
		api, err = cloudflare.NewWithAPIToken(cfg.CloudflareAPIToken)
	} else {
		api, err = cloudflare.New(cfg.CloudflareAPIKey, cfg.CloudflareEmail)
	}
	if err != nil {
		log.Fatalf("Failed to create Cloudflare API client: %v", err)
	}
	
	// Initialize services
	kvClient := kv.NewClient(api, cfg.AccountID, cfg.NamespaceID)
	authService := auth.NewService(kvClient, cfg.JWTSecret, cfg.JWTExpiration)
	
	server := &Server{
		authService: authService,
		kvClient:    kvClient,
	}
	
	// Setup routes
	http.HandleFunc("/register", server.handleRegister)
	http.HandleFunc("/login", server.handleLogin)
	http.HandleFunc("/validate", server.handleValidate)
	http.HandleFunc("/user", server.authMiddleware(server.handleGetUser))
	http.HandleFunc("/health", server.handleHealth)
	
	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	user, err := s.authService.Register(r.Context(), &req)
	if err != nil {
		s.handleError(w, err)
		return
	}
	
	s.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    user.ToUserInfo(),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	loginResp, err := s.authService.Login(r.Context(), &req)
	if err != nil {
		s.handleError(w, err)
		return
	}
	
	s.respondJSON(w, http.StatusOK, loginResp)
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	claims, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		s.handleError(w, err)
		return
	}
	
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":   true,
		"user_id": claims.UserID,
		"email":   claims.Email,
	})
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id").(string)
	
	user, err := s.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		s.handleError(w, err)
		return
	}
	
	s.respondJSON(w, http.StatusOK, user.ToUserInfo())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// authMiddleware validates JWT token and adds user info to context
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		
		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		
		claims, err := s.authService.ValidateToken(parts[1])
		if err != nil {
			s.handleError(w, err)
			return
		}
		
		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) handleError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if ok := apperrors.As(err, &appErr); ok {
		s.respondJSON(w, appErr.Code, map[string]string{
			"error": appErr.Message,
		})
		return
	}
	
	s.respondJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	})
}

func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
