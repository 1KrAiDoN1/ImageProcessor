package config

import (
	"fmt"
	"imageprocessor/backend/pkg/lib/logger/zaplogger"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func LoadServiceConfig(log *zap.Logger, configPath, dbPasswordPath string) (ServiceConfig, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		log.Error("Failed to Read config", zaplogger.Err(err))
		return ServiceConfig{}, err
	}

	var serviceConfig ServiceConfig
	if err := v.Unmarshal(&serviceConfig); err != nil {
		log.Error("Failed to Unmarshal config", zaplogger.Err(err))
		return ServiceConfig{}, err
	}
	dbConnStr, err := serviceConfig.DSN(log, dbPasswordPath)
	if err != nil {
		log.Error("Error generating DSN for database connection", zaplogger.Err(err))
		return ServiceConfig{}, err
	}
	serviceConfig.DbConfig.DBConn = dbConnStr
	log.Info("Config", zap.Any("serviceConfig", serviceConfig))
	return serviceConfig, nil
}

func (d ServiceConfig) DSN(log *zap.Logger, dbPasswordPath string) (string, error) {
	password := os.Getenv(dbPasswordPath)
	if password == "" {
		return "", fmt.Errorf("environment variable %s is not set", dbPasswordPath)
	}

	return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		d.DbConfig.Driver, d.DbConfig.User, password, d.DbConfig.Host, d.DbConfig.Port, d.DbConfig.DBName), nil
}

type ServiceConfig struct {
	Server             ServerConfig       `mapstructure:"server"`
	DbConfig           DBConfig           `mapstructure:"database"`
	BrokerConfig       BrokerConfig       `mapstructure:"broker"`
	WorkerConfig       WorkerConfig       `mapstructure:"worker"`
	CloudStorageConfig CloudStorageConfig `mapstucture:"cloud"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	User   string `yaml:"user"`
	DBName string `yaml:"dbname"`
	DBConn string
}

type ServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
}

type BrokerConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	// Group_id string   `yaml:"group_id"`
}

type WorkerConfig struct {
	NumWorkers int `yaml:"numWorkers"`
}

type CloudStorageConfig struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}
