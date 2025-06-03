package config

import (
	_ "embed"
	"log"
	"strings"

	"github.com/spf13/viper"
)

//go:embed config.local.yaml
var configData []byte

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Beta     BetaConfig     `mapstructure:"beta"`
	OCR      OCRConfig      `mapstructure:"ocr"`
	Log      LogConfig      `mapstructure:"log"`
	Trace    TraceConfig    `mapstructure:"trace"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type BetaConfig struct {
	API          BetaAPIConfig          `mapstructure:"api"`
	ModelVersion BetaModelVersionConfig `mapstructure:"model_version"`
}

type BetaAPIConfig struct {
	Overall      string `mapstructure:"overall"`
	Fluency      string `mapstructure:"fluency"`
	WordSentence string `mapstructure:"word_sentence"`
	Expression   string `mapstructure:"expression"`
	Suggestion   string `mapstructure:"suggestion"`
	Paragraph    string `mapstructure:"paragraph"`
	GrammarInfo  string `mapstructure:"grammar_info"`
	EssayInfo    string `mapstructure:"essay_info"`
}

type BetaModelVersionConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type OCRConfig struct {
	DefaultProvider string `mapstructure:"default_provider"`
	BeeAPI          string `mapstructure:"bee_api"`
	XAppKey         string `mapstructure:"x_app_key"`
	XAppSecret      string `mapstructure:"x_app_secret"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type TraceConfig struct {
	ServiceName string `mapstructure:"service_name"`
	Endpoint    string `mapstructure:"endpoint"`
}

func Load() *Config {
	viper.SetConfigType("yaml")

	if err := viper.ReadConfig(strings.NewReader(string(configData))); err != nil {
		log.Fatal("Failed to parse config:", err)
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
