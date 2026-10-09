package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	CORS       CORSConfig       `mapstructure:"cors"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Pagination PaginationConfig `mapstructure:"pagination"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Seed       SeedConfig       `mapstructure:"seed"`
}

type AppConfig struct {
	Name         string `mapstructure:"name"`
	Env          string `mapstructure:"env"`
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	LogLevel        string `mapstructure:"log_level"`
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host + " port=" + strconv.Itoa(d.Port) + " user=" + d.User + " password=" + d.Password + " dbname=" + d.Name + " sslmode=" + d.SSLMode
}

type JWTConfig struct {
	Secret            string `mapstructure:"secret"`
	AccessTokenExpiry int    `mapstructure:"access_token_expiry"`
	Issuer            string `mapstructure:"issuer"`
}

type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
}

type RateLimitConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	MaxRequests   int  `mapstructure:"max_requests"`
	WindowSeconds int  `mapstructure:"window_seconds"`
}

type PaginationConfig struct {
	DefaultLimit int `mapstructure:"default_limit"`
	MaxLimit     int `mapstructure:"max_limit"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type SeedConfig struct {
	AdminEmail     string `mapstructure:"admin_email"`
	AdminPassword  string `mapstructure:"admin_password"`
	AdminFirstName string `mapstructure:"admin_first_name"`
	AdminLastName  string `mapstructure:"admin_last_name"`

	CompanyLegalName string `mapstructure:"company_legal_name"`
	CompanyTradeName string `mapstructure:"company_trade_name"`
	CompanyTaxID     string `mapstructure:"company_tax_id"`

	CompanyAdminEmail     string `mapstructure:"company_admin_email"`
	CompanyAdminPassword  string `mapstructure:"company_admin_password"`
	CompanyAdminFirstName string `mapstructure:"company_admin_first_name"`
	CompanyAdminLastName  string `mapstructure:"company_admin_last_name"`
}

var Module = fx.Provide(NewConfig)

func NewConfig() (*Config, error) {
	loadDotEnv(".env")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("SIGIF")

	// Explicitly bind env vars for critical settings.
	_ = viper.BindEnv("app.env", "SIGIF_APP_ENV")
	_ = viper.BindEnv("app.port", "SIGIF_APP_PORT")
	_ = viper.BindEnv("app.host", "SIGIF_APP_HOST")
	_ = viper.BindEnv("database.host", "SIGIF_DATABASE_HOST")
	_ = viper.BindEnv("database.port", "SIGIF_DATABASE_PORT")
	_ = viper.BindEnv("database.user", "SIGIF_DATABASE_USER")
	_ = viper.BindEnv("database.password", "SIGIF_DATABASE_PASSWORD")
	_ = viper.BindEnv("database.name", "SIGIF_DATABASE_NAME")
	_ = viper.BindEnv("database.ssl_mode", "SIGIF_DATABASE_SSL_MODE")
	_ = viper.BindEnv("jwt.secret", "SIGIF_JWT_SECRET")
	_ = viper.BindEnv("jwt.access_token_expiry", "SIGIF_JWT_ACCESS_TOKEN_EXPIRY")
	_ = viper.BindEnv("jwt.issuer", "SIGIF_JWT_ISSUER")
	_ = viper.BindEnv("seed.admin_email", "SIGIF_SEED_ADMIN_EMAIL")
	_ = viper.BindEnv("seed.admin_password", "SIGIF_SEED_ADMIN_PASSWORD")
	_ = viper.BindEnv("seed.admin_first_name", "SIGIF_SEED_ADMIN_FIRST_NAME")
	_ = viper.BindEnv("seed.admin_last_name", "SIGIF_SEED_ADMIN_LAST_NAME")
	_ = viper.BindEnv("seed.company_legal_name", "SIGIF_SEED_COMPANY_LEGAL_NAME")
	_ = viper.BindEnv("seed.company_trade_name", "SIGIF_SEED_COMPANY_TRADE_NAME")
	_ = viper.BindEnv("seed.company_tax_id", "SIGIF_SEED_COMPANY_TAX_ID")
	_ = viper.BindEnv("seed.company_admin_email", "SIGIF_SEED_COMPANY_ADMIN_EMAIL")
	_ = viper.BindEnv("seed.company_admin_password", "SIGIF_SEED_COMPANY_ADMIN_PASSWORD")
	_ = viper.BindEnv("seed.company_admin_first_name", "SIGIF_SEED_COMPANY_ADMIN_FIRST_NAME")
	_ = viper.BindEnv("seed.company_admin_last_name", "SIGIF_SEED_COMPANY_ADMIN_LAST_NAME")

	// Config file is optional.
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadDotEnv loads KEY=VALUE pairs from a .env file into the process
// environment without overriding variables that are already set. It only
// recognizes simple lines (comments and empty lines are ignored).
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
