package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	Env  string
	Port int

	DatabaseURL     string
	JWTSecret       string
	PhoneHashPepper string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	VoiceProvider string
	SarvamAPIKey  string
	Sarvam        SarvamConfig
	Bhashini      BhashiniConfig

	LLMProvider string
	Gemini      GeminiConfig
	OpenAI      OpenAIConfig

	SMSProvider  string
	MSG91AuthKey string
	MSG91Sender  string

	WhatsAppProvider string
	WhatsAppAPIKey   string

	ServerMode string

	DataDir     string
	PDFDir      string
	FontDir     string
	MaxAudioSec int
	MaxAudioMB  int

	AllowedOrigins []string

	RateLimit RateLimitConfig
}

type BhashiniConfig struct {
	APIKey  string
	UserID  string
	BaseURL string
}

type SarvamConfig struct {
	STTModel string
	TTSModel string
	Speaker  string
}

type GeminiConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type RateLimitConfig struct {
	AnonPerMinute        int
	AnonPerHour          int
	CitizenPerMinute     int
	CitizenPerHour       int
	VoicePerMinute       int
	VoicePerHour         int
	OTPPerPhone10Min     int
	OTPPerPhoneDay       int
	HelperSessionsPerDay int
}

// Load reads configuration from the environment (and an optional .env file).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:             getEnv("APP_ENV", "development"),
		Port:            getEnvInt("PORT", 8080),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://sahayak:sahayak@localhost:5432/sahayak?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		PhoneHashPepper: getEnv("PHONE_HASH_PEPPER", ""),
		AccessTokenTTL:  time.Duration(getEnvInt("ACCESS_TOKEN_TTL_SEC", 900)) * time.Second,
		RefreshTokenTTL: time.Duration(getEnvInt("REFRESH_TOKEN_TTL_SEC", 2592000)) * time.Second,
		VoiceProvider:   strings.ToLower(getEnv("VOICE_PROVIDER", "sarvam")),
		SarvamAPIKey:    getEnv("SARVAM_API_KEY", ""),
		Sarvam: SarvamConfig{
			STTModel: getEnv("SARVAM_STT_MODEL", "saaras:v4"),
			TTSModel: getEnv("SARVAM_TTS_MODEL", "bulbul:v3"),
			Speaker:  getEnv("SARVAM_TTS_SPEAKER", "priya"),
		},
		Bhashini: BhashiniConfig{
			APIKey:  getEnv("BHASHINI_API_KEY", ""),
			UserID:  getEnv("BHASHINI_USER_ID", ""),
			BaseURL: getEnv("BHASHINI_BASE_URL", "https://dhruva-api.bhashini.gov.in"),
		},
		LLMProvider: strings.ToLower(getEnv("LLM_PROVIDER", "gemini")),
		Gemini: GeminiConfig{
			APIKey:  getEnv("GEMINI_API_KEY", ""),
			Model:   getEnv("GEMINI_MODEL", "gemini-3.5-flash-lite"),
			BaseURL: getEnv("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
		},
		OpenAI: OpenAIConfig{
			APIKey:  getEnv("OPENAI_API_KEY", ""),
			Model:   getEnv("OPENAI_MODEL", "gpt-4o-mini"),
			BaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		},
		SMSProvider:      strings.ToLower(getEnv("SMS_PROVIDER", "msg91")),
		MSG91AuthKey:     getEnv("MSG91_AUTH_KEY", ""),
		MSG91Sender:      getEnv("MSG91_SENDER", "SAHAYK"),
		WhatsAppProvider: strings.ToLower(getEnv("WHATSAPP_PROVIDER", "gupshup")),
		WhatsAppAPIKey:   getEnv("WHATSAPP_API_KEY", ""),
		ServerMode:       strings.ToLower(getEnv("SERVER_MODE", "all")),
		DataDir:          getEnv("DATA_DIR", "./data"),
		PDFDir:           getEnv("PDF_DIR", "./data/pdfs"),
		FontDir:          getEnv("FONT_DIR", "./fonts"),
		MaxAudioSec:      getEnvInt("MAX_AUDIO_SECONDS", 60),
		MaxAudioMB:       getEnvInt("MAX_AUDIO_MB", 5),
		AllowedOrigins:   splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "")),
		RateLimit: RateLimitConfig{
			AnonPerMinute:        getEnvInt("RATE_ANON_PER_MIN", 30),
			AnonPerHour:          getEnvInt("RATE_ANON_PER_HOUR", 200),
			CitizenPerMinute:     getEnvInt("RATE_CITIZEN_PER_MIN", 60),
			CitizenPerHour:       getEnvInt("RATE_CITIZEN_PER_HOUR", 500),
			VoicePerMinute:       getEnvInt("RATE_VOICE_PER_MIN", 10),
			VoicePerHour:         getEnvInt("RATE_VOICE_PER_HOUR", 100),
			OTPPerPhone10Min:     getEnvInt("RATE_OTP_PER_10MIN", 3),
			OTPPerPhoneDay:       getEnvInt("RATE_OTP_PER_DAY", 10),
			HelperSessionsPerDay: getEnvInt("RATE_HELPER_SESSIONS_PER_DAY", 50),
		},
	}

	if cfg.JWTSecret == "" {
		if cfg.Env == "production" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production")
		}
		cfg.JWTSecret = "dev-insecure-jwt-secret-change-me"
	}
	if cfg.PhoneHashPepper == "" {
		if cfg.Env == "production" {
			return nil, fmt.Errorf("PHONE_HASH_PEPPER must be set in production")
		}
		cfg.PhoneHashPepper = "dev-insecure-pepper-change-me"
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
