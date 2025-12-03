package jwt

import (
    "crypto/rand"
    "encoding/hex"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Validator interface {
    Sign(userID int64, email, role string) (string, error)
    Parse(tokenStr string) (jwt.MapClaims, error)
}

type HS256 struct {
    secret []byte
    ttl    time.Duration
    // issuer/audience фиксируем одинаково, при желании можно вынести в конфиг
}

func NewHS256(secret []byte, ttl time.Duration) *HS256 {
    return &HS256{secret: secret, ttl: ttl}
}

// Sign — для access-токенов (без jti)
func (h *HS256) Sign(userID int64, email, role string) (string, error) {
    now := time.Now()
    claims := jwt.MapClaims{
        "sub":   userID,
        "email": email,
        "role":  role,
        "iat":   now.Unix(),
        "exp":   now.Add(h.ttl).Unix(),
        "iss":   "pz10-auth",
        "audi":  "pz10-clients", // используем "audi" чтобы не конфликтовать с "aud" array
    }
    t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return t.SignedString(h.secret)
}

// SignRefresh — для refresh-токенов (с jti для отзыва)
func (h *HS256) SignRefresh(userID int64, email, role string) (string, string, error) {
    now := time.Now()
    jti := newJTI()
    claims := jwt.MapClaims{
        "sub":   userID,
        "email": email,
        "role":  role,
        "iat":   now.Unix(),
        "exp":   now.Add(h.ttl).Unix(),
        "iss":   "pz10-auth",
        "audi":  "pz10-clients",
        "jti":   jti,
    }
    t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := t.SignedString(h.secret)
    return signed, jti, err
}

func (h *HS256) Parse(tokenStr string) (jwt.MapClaims, error) {
    t, err := jwt.Parse(tokenStr,
        func(t *jwt.Token) (any, error) { return h.secret, nil },
        jwt.WithValidMethods([]string{"HS256"}),
        jwt.WithIssuer("pz10-auth"),
        jwt.WithAudience("pz10-clients"),
    )
    if err != nil || !t.Valid {
        return nil, err
    }
    return t.Claims.(jwt.MapClaims), nil
}

func newJTI() string {
    var b [16]byte
    _, _ = rand.Read(b[:])
    return hex.EncodeToString(b[:])
}
