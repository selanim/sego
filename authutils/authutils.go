package authutils

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Aina ya password hashing algorithm
type HashAlgorithm string

const (
	BCrypt   HashAlgorithm = "bcrypt"
	Argon2ID HashAlgorithm = "argon2id"
	PBKDF2   HashAlgorithm = "pbkdf2"
)

// Mipangilio ya password hashing
type PasswordConfig struct {
	Algorithm  HashAlgorithm
	Cost       int    // Kwa BCrypt
	Time       uint32 // Kwa Argon2
	Memory     uint32 // Kwa Argon2
	Threads    uint8  // Kwa Argon2
	KeyLength  uint32 // Urefu wa key
	SaltLength int    // Urefu wa salt
	Iterations int    // Kwa PBKDF2
}

// TokenConfig ina mipangilio ya JWT token
type TokenConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
	Audience           []string
}

// UserData ina data ya user inayohitajika kwa authentication
type UserData struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	Salt         string // Ikiwa algorithm inahitaji salt tofauti
	Algorithm    HashAlgorithm
	IsActive     bool
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Roles        []string
	Permissions  []string
}

// TokenDetails inaelezea token zilizogenerate
type TokenDetails struct {
	AccessToken      string
	RefreshToken     string
	AccessUUID       string
	RefreshUUID      string
	AccessExpiresAt  int64
	RefreshExpiresAt int64
}

// Claims ni JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// AuthError ni custom error ya authentication
type AuthError struct {
	Code    string
	Message string
	Err     error
}

func (e *AuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAuthError(code, message string, err error) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Default password configurations
var DefaultPasswordConfigs = map[HashAlgorithm]PasswordConfig{
	BCrypt: {
		Algorithm: BCrypt,
		Cost:      12,
	},
	Argon2ID: {
		Algorithm:  Argon2ID,
		Time:       3,
		Memory:     64 * 1024, // 64MB
		Threads:    4,
		KeyLength:  32,
		SaltLength: 16,
	},
	PBKDF2: {
		Algorithm:  PBKDF2,
		Iterations: 310000, // Recommended by OWASP
		SaltLength: 16,
		KeyLength:  32,
	},
}

// AuthUtils ni struct kuu ya authentication utilities
type AuthUtils struct {
	TokenConfig    TokenConfig
	PasswordConfig PasswordConfig
	TokenStore     TokenStore // Optional: kwa token management
}

// TokenStore interface kwa token storage (Redis, database, n.k)
type TokenStore interface {
	SetToken(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetToken(ctx context.Context, key string) (string, error)
	DeleteToken(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// NewAuthUtils inainitialize AuthUtils na default configurations
func NewAuthUtils(tokenConfig TokenConfig) *AuthUtils {
	return &AuthUtils{
		TokenConfig:    tokenConfig,
		PasswordConfig: DefaultPasswordConfigs[Argon2ID], // Default to Argon2ID
	}
}

// NewAuthUtilsWithConfig inainitialize na custom configurations
func NewAuthUtilsWithConfig(tokenConfig TokenConfig, passwordConfig PasswordConfig) *AuthUtils {
	return &AuthUtils{
		TokenConfig:    tokenConfig,
		PasswordConfig: passwordConfig,
	}
}

// GenerateRandomBytes inatengeneza random bytes kwa ajili ya salts na tokens
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, NewAuthError("RANDOM_ERROR", "Failed to generate random bytes", err)
	}
	return b, nil
}

// GenerateSalt inatengeneza salt ya password hashing
func (au *AuthUtils) GenerateSalt() (string, error) {
	saltLength := au.PasswordConfig.SaltLength
	if saltLength == 0 {
		saltLength = 16 // Default
	}

	salt, err := GenerateRandomBytes(saltLength)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(salt), nil
}

// HashPassword inahash password kwa kutumia algorithm iliyochaguliwa
func (au *AuthUtils) HashPassword(password string) (string, string, error) {
	if password == "" {
		return "", "", NewAuthError("VALIDATION_ERROR", "Password cannot be empty", nil)
	}

	switch au.PasswordConfig.Algorithm {
	case BCrypt:
		return au.hashWithBCrypt(password)
	case Argon2ID:
		return au.hashWithArgon2(password)
	case PBKDF2:
		return au.hashWithPBKDF2(password)
	default:
		return "", "", NewAuthError("ALGORITHM_ERROR", "Unsupported hashing algorithm", nil)
	}
}

// hashWithBCrypt inahash password kwa kutumia BCrypt
func (au *AuthUtils) hashWithBCrypt(password string) (string, string, error) {
	cost := au.PasswordConfig.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", "", NewAuthError("BCRYPT_ERROR", "Failed to hash password with BCrypt", err)
	}

	return string(hashBytes), "", nil // BCrypt haina salt tofauti
}

// hashWithArgon2 inahash password kwa kutumia Argon2
func (au *AuthUtils) hashWithArgon2(password string) (string, string, error) {
	config := au.PasswordConfig

	// Generate salt
	salt, err := GenerateRandomBytes(config.SaltLength)
	if err != nil {
		return "", "", err
	}

	// Hash password
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLength,
	)

	// Encode hash and salt
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	params := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		config.Memory,
		config.Time,
		config.Threads,
		saltEncoded,
		hashEncoded,
	)

	return params, saltEncoded, nil
}

// hashWithPBKDF2 inahash password kwa kutumia PBKDF2
func (au *AuthUtils) hashWithPBKDF2(password string) (string, string, error) {
	config := au.PasswordConfig

	// Generate salt
	salt, err := GenerateRandomBytes(config.SaltLength)
	if err != nil {
		return "", "", err
	}

	// Hash password using PBKDF2 with HMAC-SHA256
	hash := pbkdf2.Key([]byte(password), salt, config.Iterations, int(config.KeyLength), sha256.New)

	// Encode hash and salt
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $pbkdf2-sha256$i=310000$salt$hash
	params := fmt.Sprintf("$pbkdf2-sha256$i=%d$%s$%s",
		config.Iterations,
		saltEncoded,
		hashEncoded,
	)

	return params, saltEncoded, nil
}

// VerifyPassword inaangalia kama password inafanana na hash
func (au *AuthUtils) VerifyPassword(password, hash, salt string) (bool, error) {
	if password == "" || hash == "" {
		return false, NewAuthError("VALIDATION_ERROR", "Password and hash cannot be empty", nil)
	}

	// Detect algorithm from hash format
	algorithm := au.detectAlgorithmFromHash(hash)

	switch algorithm {
	case BCrypt:
		return au.verifyBCrypt(password, hash)
	case Argon2ID:
		return au.verifyArgon2(password, hash)
	case PBKDF2:
		return au.verifyPBKDF2(password, hash)
	default:
		// Assume BCrypt ikiwa haijatambulika
		return au.verifyBCrypt(password, hash)
	}
}

// detectAlgorithmFromHash inatambua algorithm kutoka kwa format ya hash
func (au *AuthUtils) detectAlgorithmFromHash(hash string) HashAlgorithm {
	if strings.HasPrefix(hash, "$argon2id$") {
		return Argon2ID
	} else if strings.HasPrefix(hash, "$pbkdf2-") {
		return PBKDF2
	} else if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$") {
		return BCrypt
	}
	return BCrypt // Default
}

// verifyBCrypt inaverify password kwa BCrypt
func (au *AuthUtils) verifyBCrypt(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, NewAuthError("BCRYPT_ERROR", "Failed to verify password", err)
	}
	return true, nil
}

// verifyArgon2 inaverify password kwa Argon2
func (au *AuthUtils) verifyArgon2(password, hash string) (bool, error) {
	// Parse hash parameters
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, NewAuthError("ARGON2_ERROR", "Invalid Argon2 hash format", nil)
	}

	// Extract parameters
	params := strings.Split(parts[3], ",")
	var memory, time uint32
	var threads uint8

	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "m":
			fmt.Sscanf(kv[1], "%d", &memory)
		case "t":
			fmt.Sscanf(kv[1], "%d", &time)
		case "p":
			fmt.Sscanf(kv[1], "%d", &threads)
		}
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, NewAuthError("DECODE_ERROR", "Failed to decode salt", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, NewAuthError("DECODE_ERROR", "Failed to decode hash", err)
	}

	// Compute hash with same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		threads,
		uint32(len(expectedHash)),
	)

	// Compare hashes with constant time comparison
	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}

// verifyPBKDF2 inaverify password kwa PBKDF2
func (au *AuthUtils) verifyPBKDF2(password, hash string) (bool, error) {
	// Parse hash parameters
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		return false, NewAuthError("PBKDF2_ERROR", "Invalid PBKDF2 hash format", nil)
	}

	// Extract iterations
	var iterations int
	fmt.Sscanf(strings.TrimPrefix(parts[2], "i="), "%d", &iterations)

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, NewAuthError("DECODE_ERROR", "Failed to decode salt", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, NewAuthError("DECODE_ERROR", "Failed to decode hash", err)
	}

	// Compute hash with same parameters
	computedHash := pbkdf2.Key(
		[]byte(password),
		salt,
		iterations,
		len(expectedHash),
		sha256.New,
	)

	// Compare hashes with constant time comparison
	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}

// GenerateTokens inatengeneza access na refresh tokens
func (au *AuthUtils) GenerateTokens(userData *UserData) (*TokenDetails, error) {
	if userData == nil {
		return nil, NewAuthError("VALIDATION_ERROR", "User data cannot be nil", nil)
	}

	// Generate unique IDs for tokens
	accessUUID, err := GenerateRandomBytes(16)
	if err != nil {
		return nil, err
	}

	refreshUUID, err := GenerateRandomBytes(16)
	if err != nil {
		return nil, err
	}

	accessUUIDStr := hex.EncodeToString(accessUUID)
	refreshUUIDStr := hex.EncodeToString(refreshUUID)

	now := time.Now()

	// Access token claims
	accessClaims := &Claims{
		UserID:      userData.ID,
		Username:    userData.Username,
		Email:       userData.Email,
		Roles:       userData.Roles,
		Permissions: userData.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(au.TokenConfig.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    au.TokenConfig.Issuer,
			Audience:  au.TokenConfig.Audience,
			ID:        accessUUIDStr,
		},
	}

	// Refresh token claims
	refreshClaims := &Claims{
		UserID:   userData.ID,
		Username: userData.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(au.TokenConfig.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    au.TokenConfig.Issuer,
			Audience:  au.TokenConfig.Audience,
			ID:        refreshUUIDStr,
		},
	}

	// Create tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Sign tokens
	accessTokenString, err := accessToken.SignedString([]byte(au.TokenConfig.SecretKey))
	if err != nil {
		return nil, NewAuthError("TOKEN_ERROR", "Failed to sign access token", err)
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(au.TokenConfig.SecretKey))
	if err != nil {
		return nil, NewAuthError("TOKEN_ERROR", "Failed to sign refresh token", err)
	}

	return &TokenDetails{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshTokenString,
		AccessUUID:       accessUUIDStr,
		RefreshUUID:      refreshUUIDStr,
		AccessExpiresAt:  now.Add(au.TokenConfig.AccessTokenExpiry).Unix(),
		RefreshExpiresAt: now.Add(au.TokenConfig.RefreshTokenExpiry).Unix(),
	}, nil
}

// ValidateToken ina validate JWT token
func (au *AuthUtils) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, NewAuthError("VALIDATION_ERROR", "Token cannot be empty", nil)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, NewAuthError("TOKEN_ERROR", "Unexpected signing method", nil)
		}
		return []byte(au.TokenConfig.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, NewAuthError("TOKEN_EXPIRED", "Token has expired", err)
		}
		return nil, NewAuthError("TOKEN_ERROR", "Invalid token", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Additional validation
		if !au.validateClaims(claims) {
			return nil, NewAuthError("CLAIMS_ERROR", "Invalid token claims", nil)
		}
		return claims, nil
	}

	return nil, NewAuthError("TOKEN_ERROR", "Invalid token claims", nil)
}

// validateClaims ina validate additional claims
func (au *AuthUtils) validateClaims(claims *Claims) bool {
	// Check issuer
	if claims.Issuer != "" && au.TokenConfig.Issuer != "" && claims.Issuer != au.TokenConfig.Issuer {
		return false
	}

	// Check audience
	if len(claims.Audience) > 0 && len(au.TokenConfig.Audience) > 0 {
		audienceValid := false
		for _, aud := range claims.Audience {
			for _, expectedAud := range au.TokenConfig.Audience {
				if aud == expectedAud {
					audienceValid = true
					break
				}
			}
			if audienceValid {
				break
			}
		}
		if !audienceValid {
			return false
		}
	}

	return true
}

// ExtractTokenFromRequest ina extract token kutoka HTTP request
func (au *AuthUtils) ExtractTokenFromRequest(r *http.Request) string {
	// Check Authorization header
	bearerToken := r.Header.Get("Authorization")
	if bearerToken != "" {
		// Format: Bearer <token>
		parts := strings.Split(bearerToken, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Check query parameter
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	// Check cookie
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// RefreshTokens ina generate tokens mpya kutokana na refresh token
func (au *AuthUtils) RefreshTokens(refreshToken string, userData *UserData) (*TokenDetails, error) {
	// Validate refresh token
	claims, err := au.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verify it's a refresh token (optional: check token type in claims)
	// Generate new tokens
	return au.GenerateTokens(userData)
}

// ValidatePasswordStrength ina validate password strength
func (au *AuthUtils) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return NewAuthError("PASSWORD_WEAK", "Password must be at least 8 characters long", nil)
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return NewAuthError("PASSWORD_WEAK", "Password must contain at least one uppercase letter", nil)
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return NewAuthError("PASSWORD_WEAK", "Password must contain at least one lowercase letter", nil)
	}

	// Check for at least one digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return NewAuthError("PASSWORD_WEAK", "Password must contain at least one digit", nil)
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)
	if !hasSpecial {
		return NewAuthError("PASSWORD_WEAK", "Password must contain at least one special character", nil)
	}

	// Check common passwords (simplified)
	commonPasswords := []string{
		"password", "12345678", "qwerty", "admin", "welcome",
		"password123", "123456789", "1234567", "123456",
	}

	for _, common := range commonPasswords {
		if strings.ToLower(password) == common {
			return NewAuthError("PASSWORD_COMMON", "Password is too common", nil)
		}
	}

	return nil
}

// GenerateOTP ina generate One-Time Password
func (au *AuthUtils) GenerateOTP(length int) (string, error) {
	if length < 4 || length > 8 {
		length = 6 // Default OTP length
	}

	// Generate random number
	max := 1
	for i := 0; i < length; i++ {
		max *= 10
	}

	randomBytes, err := GenerateRandomBytes(4)
	if err != nil {
		return "", err
	}

	// Convert to number within range
	randomInt := int(binary.BigEndian.Uint32(randomBytes)) % max
	if randomInt < 0 {
		randomInt = -randomInt
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, randomInt), nil
}

// GenerateAPIKey ina generate API key
func (au *AuthUtils) GenerateAPIKey(prefix string, length int) (string, error) {
	if length < 16 {
		length = 32 // Default length
	}

	keyBytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	apiKey := hex.EncodeToString(keyBytes)

	if prefix != "" {
		apiKey = prefix + "_" + apiKey
	}

	return apiKey, nil
}

// GenerateSecureToken ina generate secure random token
func (au *AuthUtils) GenerateSecureToken(length int) (string, error) {
	if length < 16 {
		length = 32 // Default length
	}

	tokenBytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	// Use URL-safe base64 encoding
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// HashToken ina hash token kwa ajili ya storage
func (au *AuthUtils) HashToken(token string) (string, error) {
	// Use SHA256 for token hashing
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}

// VerifyTokenHash ina verify token hash
func (au *AuthUtils) VerifyTokenHash(token, hash string) bool {
	computedHash := sha256.Sum256([]byte(token))
	computedHashStr := hex.EncodeToString(computedHash[:])

	return subtle.ConstantTimeCompare([]byte(computedHashStr), []byte(hash)) == 1
}

// Middleware ya authentication kwa HTTP handlers
func (au *AuthUtils) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := au.ExtractTokenFromRequest(r)
		if token == "" {
			au.writeError(w, http.StatusUnauthorized, "Authorization token required")
			return
		}

		claims, err := au.ValidateToken(token)
		if err != nil {
			au.writeError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "userClaims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleBasedMiddleware ina restrict access kulingana na roles
func (au *AuthUtils) RoleBasedMiddleware(allowedRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("userClaims").(*Claims)
			if !ok || claims == nil {
				au.writeError(w, http.StatusUnauthorized, "User claims not found")
				return
			}

			// Check if user has any of the allowed roles
			hasRole := false
			for _, userRole := range claims.Roles {
				for _, allowedRole := range allowedRoles {
					if userRole == allowedRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				au.writeError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// PermissionBasedMiddleware ina restrict access kulingana na permissions
func (au *AuthUtils) PermissionBasedMiddleware(requiredPermissions []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("userClaims").(*Claims)
			if !ok || claims == nil {
				au.writeError(w, http.StatusUnauthorized, "User claims not found")
				return
			}

			// Check if user has all required permissions
			for _, requiredPerm := range requiredPermissions {
				hasPerm := false
				for _, userPerm := range claims.Permissions {
					if userPerm == requiredPerm {
						hasPerm = true
						break
					}
				}
				if !hasPerm {
					au.writeError(w, http.StatusForbidden,
						fmt.Sprintf("Missing permission: %s", requiredPerm))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeError inaandika error response
func (au *AuthUtils) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":     true,
		"code":      statusCode,
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// Rate limiting utilities
type RateLimiter struct {
	store  TokenStore
	limit  int
	window time.Duration
}

func NewRateLimiter(store TokenStore, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		store:  store,
		limit:  limit,
		window: window,
	}
}

// CheckRateLimit ina check rate limit for a key
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, key string) (bool, int, error) {
	currentKey := fmt.Sprintf("ratelimit:%s", key)

	// Get current count
	countStr, err := rl.store.GetToken(ctx, currentKey)
	if err != nil {
		// Key doesn't exist, create it
		err = rl.store.SetToken(ctx, currentKey, "1", rl.window)
		return true, 1, err
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 0
	}

	if count >= rl.limit {
		return false, count, nil
	}

	// Increment count
	count++
	err = rl.store.SetToken(ctx, currentKey, strconv.Itoa(count), rl.window)
	return true, count, err
}

// Password expiration utilities
func (au *AuthUtils) ShouldChangePassword(lastChanged time.Time, maxAgeDays int) bool {
	if maxAgeDays <= 0 {
		return false // No expiration
	}

	expirationDate := lastChanged.Add(time.Duration(maxAgeDays) * 24 * time.Hour)
	return time.Now().After(expirationDate)
}

// Session management
func (au *AuthUtils) CreateSession(ctx context.Context, userID string, tokenDetails *TokenDetails) error {
	if au.TokenStore == nil {
		return NewAuthError("STORAGE_ERROR", "Token store not configured", nil)
	}

	// Store access token
	accessKey := fmt.Sprintf("access:%s", tokenDetails.AccessUUID)
	err := au.TokenStore.SetToken(ctx, accessKey, userID,
		time.Until(time.Unix(tokenDetails.AccessExpiresAt, 0)))
	if err != nil {
		return err
	}

	// Store refresh token
	refreshKey := fmt.Sprintf("refresh:%s", tokenDetails.RefreshUUID)
	err = au.TokenStore.SetToken(ctx, refreshKey, userID,
		time.Until(time.Unix(tokenDetails.RefreshExpiresAt, 0)))

	return err
}

func (au *AuthUtils) ValidateSession(ctx context.Context, tokenUUID string) (string, error) {
	if au.TokenStore == nil {
		return "", NewAuthError("STORAGE_ERROR", "Token store not configured", nil)
	}

	key := fmt.Sprintf("access:%s", tokenUUID)
	userID, err := au.TokenStore.GetToken(ctx, key)
	if err != nil {
		return "", NewAuthError("SESSION_ERROR", "Invalid or expired session", err)
	}

	return userID, nil
}

func (au *AuthUtils) RevokeSession(ctx context.Context, tokenUUID string) error {
	if au.TokenStore == nil {
		return NewAuthError("STORAGE_ERROR", "Token store not configured", nil)
	}

	key := fmt.Sprintf("access:%s", tokenUUID)
	return au.TokenStore.DeleteToken(ctx, key)
}

func (au *AuthUtils) RevokeAllUserSessions(ctx context.Context, userID string) error {
	// Implementation depends on storage strategy
	// This could use a scan pattern in Redis or database query
	return nil
}

// Utility functions
func (au *AuthUtils) GetPasswordStrengthScore(password string) int {
	score := 0

	// Length score
	if len(password) >= 8 {
		score += 1
	}
	if len(password) >= 12 {
		score += 1
	}
	if len(password) >= 16 {
		score += 1
	}

	// Character variety score
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score += 1
	}
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		score += 1
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		score += 1
	}
	if regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(password) {
		score += 1
	}

	// Bonus for no dictionary words
	if !au.containsDictionaryWords(password) {
		score += 1
	}

	return score
}

func (au *AuthUtils) containsDictionaryWords(password string) bool {
	// Simplified check - in production, use a proper dictionary
	commonWords := []string{"password", "admin", "welcome", "qwerty", "123456"}
	lowerPassword := strings.ToLower(password)

	for _, word := range commonWords {
		if strings.Contains(lowerPassword, word) {
			return true
		}
	}

	return false
}

// Generate recovery token
func (au *AuthUtils) GenerateRecoveryToken() (string, error) {
	return au.GenerateSecureToken(32)
}

// Validate recovery token expiry
func (au *AuthUtils) ValidateRecoveryTokenExpiry(createdAt time.Time, expiryHours int) bool {
	expiryTime := createdAt.Add(time.Duration(expiryHours) * time.Hour)
	return time.Now().Before(expiryTime)
}

// Mask email for display
func (au *AuthUtils) MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return "***@" + domain
	}

	masked := localPart[:2] + "***"
	if len(localPart) > 5 {
		masked += localPart[len(localPart)-1:]
	}

	return masked + "@" + domain
}

// Sanitize input
func (au *AuthUtils) SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	input = strings.TrimSpace(input)

	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	input = re.ReplaceAllString(input, "")

	// Remove control characters
	re = regexp.MustCompile(`[\x00-\x1F\x7F]`)
	input = re.ReplaceAllString(input, "")

	return input
}

// Note: Make sure to add required imports at the top
