package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DBSettings      DBSettings      `yaml:"DBSettings"`
	ProgramSettings ProgramSettings `yaml:"ProgramSettings"`
}

type DBSettings struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"` //измениться на master для k8s, позже
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
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
