package config

import (
	"strconv"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Pagination PaginationConfig `mapstructure:"pagination"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Modules   ModulesConfig   `mapstructure:"modules"`
	BusinessTypes []string    `mapstructure:"business_types"`
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

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

func (r RedisConfig) Addr() string {
	return r.Host + ":" + strconv.Itoa(r.Port)
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpiry  int    `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry int    `mapstructure:"refresh_token_expiry"`
	Issuer             string `mapstructure:"issuer"`
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

type ModulesConfig struct {
	Enabled []string `mapstructure:"enabled"`
}

var Module = fx.Provide(NewConfig)

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("SIGIF")

	// Explicitly bind env vars for critical settings
	_ = viper.BindEnv("database.host", "SIGIF_DATABASE_HOST")
	_ = viper.BindEnv("database.port", "SIGIF_DATABASE_PORT")
	_ = viper.BindEnv("database.user", "SIGIF_DATABASE_USER")
	_ = viper.BindEnv("database.password", "SIGIF_DATABASE_PASSWORD")
	_ = viper.BindEnv("database.name", "SIGIF_DATABASE_NAME")
	_ = viper.BindEnv("database.ssl_mode", "SIGIF_DATABASE_SSL_MODE")
	_ = viper.BindEnv("redis.host", "SIGIF_REDIS_HOST")
	_ = viper.BindEnv("redis.port", "SIGIF_REDIS_PORT")
	_ = viper.BindEnv("redis.password", "SIGIF_REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "SIGIF_REDIS_DB")
	_ = viper.BindEnv("redis.pool_size", "SIGIF_REDIS_POOL_SIZE")
	_ = viper.BindEnv("redis.min_idle_conns", "SIGIF_REDIS_MIN_IDLE_CONNS")
	_ = viper.BindEnv("jwt.secret", "SIGIF_JWT_SECRET")
	_ = viper.BindEnv("jwt.access_token_expiry", "SIGIF_JWT_ACCESS_TOKEN_EXPIRY")
	_ = viper.BindEnv("jwt.refresh_token_expiry", "SIGIF_JWT_REFRESH_TOKEN_EXPIRY")
	_ = viper.BindEnv("jwt.issuer", "SIGIF_JWT_ISSUER")
	_ = viper.BindEnv("app.env", "SIGIF_APP_ENV")
	_ = viper.BindEnv("app.port", "SIGIF_APP_PORT")
	_ = viper.BindEnv("app.host", "SIGIF_APP_HOST")

	// Config file is optional
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) IsModuleEnabled(name string) bool {
	for _, m := range c.Modules.Enabled {
		if m == name {
			return true
		}
	}
	return false
}