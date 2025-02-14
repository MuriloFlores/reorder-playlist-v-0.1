package config

import (
	"github.com/spf13/viper"
	"log"
)

var EnvConfigs *envConfig

type envConfig struct {
	ClientID      string `mapstructure:"CLIENT_ID"`
	SecretKey     string `mapstructure:"SECRET_KEY"`
	SessionSecret string `mapstructure:"SESSION_SECRET"`
	MaxAge        int    `mapstructure:"MAX_AGE"`
	Prod          bool   `mapstructure:"PROD"`
	SessionName   string `mapstructure:"SESSION_NAME"`
	RedirectURL   string `mapstructure:"REDIRECT_URL"`
	DbHost        string `mapstructure:"DB_HOST"`
	DbPort        string `mapstructure:"DB_PORT"`
	DbUsername    string `mapstructure:"DB_USERNAME"`
	DbPassword    string `mapstructure:"DB_PASSWORD"`
	DbDatabase    string `mapstructure:"DB_DATABASE"`
}

func InitEnvConfig() {
	EnvConfigs = loadEnv()
}

func loadEnv() (config *envConfig) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error unmarshalling config, %s", err)
	}

	return
}
