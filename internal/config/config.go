package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Program ProgramSettings `yaml:"program"`
	WriteDB DBSettings      `yaml:"writeDB"`
	ReadDB  DBSettings      `yaml:"readDB"`
}

type DBSettings struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

type ProgramSettings struct {
	JiraURL           string `yaml:"jiraUrl"`
	ThreadCount       int    `yaml:"threadCount"`
	IssueInOneRequest int    `yaml:"issueInOneRequest"`
	MinTimeSleep      int    `yaml:"minTimeSleep"`
	MaxTimeSleep      int    `yaml:"maxTimeSleep"`
	Port              int    `yaml:"port"`
}

func LoadConfig(filename string) (*Config, error) {
	configData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("while reading config file: %w", err)
	}

	var config Config

	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("while parsing yaml: %w", err)
	}

	return &config, nil
}
