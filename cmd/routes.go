package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type SupabaseJWTClaims struct {
	Sub   string `json:"sub"`   // User ID from Supabase
	Email string `json:"email"` // User email
	Role  string `json:"role"`  // User role
	jwt.RegisteredClaims
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func (app *application) Routes() *echo.Echo {
	e := echo.New()

	e.Use(ServerHeader)

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339} ${status} ${method} ${host}${path} ${latency_human}]` + "\n",
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "https://yata-one.vercel.app"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	secured := e.Group("/v1")

	secured.Use(app.SupabaseJWTMiddleware())

	secured.GET("/todos", app.GetTodosByUserID)

	secured.POST("/todos", app.AddNewTodo)

	secured.POST("/editTodo", app.EditExistingTodo)

	secured.POST("/tag", app.AddNewTag)

	secured.POST("/addTodoTag", app.AddTagToTodo)

	secured.POST("/toggleComplete", app.ToggleTodoCompleted)

	secured.DELETE("/todos", app.DeleteTodo)

	return e
}

func (app *application) SupabaseJWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			// Verify Supabase JWT
			claims, err := app.verifySupabaseJWT(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token: "+err.Error())
			}

			// Add user info to context
			c.Set("user_id", claims.Sub)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)

			return next(c)
		}
	}
}

func (app *application) verifySupabaseJWT(tokenString string) (*SupabaseJWTClaims, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL not set")
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &SupabaseJWTClaims{})
	if err != nil {
		return nil, err
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("kid not found in token header")
	}

	publicKey, err := app.getSupabasePublicKey(supabaseURL, kid)
	if err != nil {
		return nil, err
	}

	claims := &SupabaseJWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (app *application) getSupabasePublicKey(supabaseURL, kid string) (*rsa.PublicKey, error) {
	jwksURL := fmt.Sprintf("%s/auth/v1/jwks", supabaseURL)

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return app.jwkToRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key not found")
}

func (app *application) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)

	var e int
	if len(eBytes) == 3 {
		e = int(eBytes[0])<<16 + int(eBytes[1])<<8 + int(eBytes[2])
	} else if len(eBytes) == 1 {
		e = int(eBytes[0])
	} else {
		return nil, fmt.Errorf("invalid exponent")
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

func GetUserID(c echo.Context) string {
	if userID, ok := c.Get("user_id").(string); ok {
		return userID
	}
	return ""
}

func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "TodoApi/0.1")

		return next(c)
	}
}
