package utils

import (
    "testing"

    "github.com/golang-jwt/jwt/v5"
)

func TestGenerateTokenUsesEnvSecret(t *testing.T) {
    t.Setenv("JWT_SECRET", "env-secret-for-test")

    tokenString, err := GenerateToken(7, "alex")
    if err != nil {
        t.Fatalf("GenerateToken returned error: %v", err)
    }

    parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte("env-secret-for-test"), nil
    })
    if err != nil {
        t.Fatalf("expected token signed by env secret, got error: %v", err)
    }

    claims, ok := parsed.Claims.(*Claims)
    if !ok || !parsed.Valid {
        t.Fatalf("expected valid claims, got %#v", parsed.Claims)
    }

    if claims.UserID != 7 || claims.Username != "alex" {
        t.Fatalf("unexpected claims: %+v", claims)
    }
}
