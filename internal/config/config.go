package config

import (
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type Config struct {
	Jira   JiraConfig   `mapstructure:"jira"`
	Kafka  KafkaConfig  `mapstructure:"kafka"`
	Server ServerConfig `mapstructure:"server"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type JiraConfig struct {
	Program ProgramSettings `mapstructure:"program"`
	WriteDB DBConfig        `mapstructure:"writeDB"`
	ReadDB  DBConfig        `mapstructure:"readDB"`
}

type DBConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

type ProgramSettings struct {
	JiraURL           string `mapstructure:"jiraUrl"`
	ThreadCount       int    `mapstructure:"threadCount"`
	IssueInOneRequest int    `mapstructure:"issueInOneRequest"`
	MinTimeSleep      int    `mapstructure:"minTimeSleep"`
	MaxTimeSleep      int    `mapstructure:"maxTimeSleep"`
	Port              int    `mapstructure:"port"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Error reading config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &cfg, nil
}
