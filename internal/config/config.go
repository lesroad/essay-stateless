package config

import (
	_ "embed"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// //go:embed config.local.yaml
var configData []byte

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Evaluate EvaluateConfig `mapstructure:"evaluate"`
	OCR      OCRConfig      `mapstructure:"ocr"`
	Log      LogConfig      `mapstructure:"log"`
	Trace    TraceConfig    `mapstructure:"trace"`
	Lago     LagoConfig     `mapstructure:"lago"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type EvaluateConfig struct {
	API          EvaluateAPIConfig          `mapstructure:"api"`
	ModelVersion EvaluateModelVersionConfig `mapstructure:"model_version"`
}

type EvaluateAPIConfig struct {
	Overall      string `mapstructure:"overall"`
	WordSentence string `mapstructure:"word_sentence"`
	Suggestion   string `mapstructure:"suggestion"`
	Paragraph    string `mapstructure:"paragraph"`
	GrammarInfo  string `mapstructure:"grammar_info"`
	Score        string `mapstructure:"score"`
	EssayInfo    string `mapstructure:"essay_info"`
	Polishing    string `mapstructure:"polishing"`
}

type EvaluateModelVersionConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type OCRConfig struct {
	DefaultProvider string `mapstructure:"default_provider"`
	BeeAPI          string `mapstructure:"bee_api"`
	XAppKey         string `mapstructure:"x_app_key"`
	XAppSecret      string `mapstructure:"x_app_secret"`
	// ARK 大模型配置
	ArkAPIKey  string `mapstructure:"ark_api_key"`
	ArkBaseURL string `mapstructure:"ark_base_url"`
	ArkModel   string `mapstructure:"ark_model"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type TraceConfig struct {
	ServiceName string `mapstructure:"service_name"`
	Endpoint    string `mapstructure:"endpoint"`
}

type LagoConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Enabled bool   `mapstructure:"enabled"`
}

func Load() *Config {
	viper.SetConfigType("yaml")
	if len(configData) > 0 {
		if err := viper.ReadConfig(strings.NewReader(string(configData))); err != nil {
			log.Fatal("Failed to parse config from memory:", err)
		}
	} else {
		path := os.Getenv("CONFIG_PATH")
		if path == "" {
			log.Fatal("CONFIG_PATH environment variable is not set and no config data provided")
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatal("Config file not found:", path)
		}

		viper.SetConfigFile(path)

		if err := viper.ReadInConfig(); err != nil {
			log.Fatal("Failed to read config file:", err)
		}
	}

	setDefaults()

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("Failed to unmarshal config:", err)
	}

	return &config
}

func setDefaults() {
	viper.SetDefault("server.port", ":8090")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("trace.service_name", "essay-stateless")
}
