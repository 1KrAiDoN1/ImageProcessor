package config

import (
	"fmt"
	"imageprocessor/backend/pkg/lib/logger/zaplogger"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func LoadServiceConfig(log *zap.Logger, configPath, dbPasswordPath, cloudAccessKeyPath, cloudSecretKeyPath string) (*ServiceConfig, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		log.Error("Failed to Read config", zaplogger.Err(err))
		return &ServiceConfig{}, err
	}

	serviceConfig := &ServiceConfig{}
	if err := v.Unmarshal(serviceConfig); err != nil {
		log.Error("Failed to Unmarshal config", zaplogger.Err(err))
		return &ServiceConfig{}, err
	}
	dbConnStr, err := serviceConfig.DSN(log, dbPasswordPath)
	if err != nil {
		log.Error("Error generating DSN for database connection", zaplogger.Err(err))
		return &ServiceConfig{}, err
	}

	CloudAccessKey, CloudSecretKey, err := serviceConfig.GetCloudKeys(cloudAccessKeyPath, cloudSecretKeyPath)
	if err != nil {
		log.Error("Error getting cloud storage keys", zaplogger.Err(err))
		return &ServiceConfig{}, err
	}

	serviceConfig.DbConfig.DBConn = dbConnStr
	serviceConfig.CloudStorageConfig.AccessKey = CloudAccessKey
	serviceConfig.CloudStorageConfig.SecretKey = CloudSecretKey

	log.Info("Config", zap.Any("serviceConfig", serviceConfig))
	return serviceConfig, nil
}

func (d *ServiceConfig) DSN(log *zap.Logger, dbPasswordPath string) (string, error) {
	password := os.Getenv(dbPasswordPath)
	if password == "" {
		return "", fmt.Errorf("environment variable %s is not set", dbPasswordPath)
	}

	return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		d.DbConfig.Driver, d.DbConfig.User, password, d.DbConfig.Host, d.DbConfig.Port, d.DbConfig.DBName), nil
}

func (d *ServiceConfig) GetCloudKeys(CloudAccessKeyPath, CloudSecretKeyPath string) (string, string, error) {
	accessKey := os.Getenv(CloudAccessKeyPath)
	secretKey := os.Getenv(CloudSecretKeyPath)
	if accessKey == "" || secretKey == "" {
		return "", "", fmt.Errorf("environment variables %s and/or %s are not set", CloudAccessKeyPath, CloudSecretKeyPath)
	}
	return accessKey, secretKey, nil
}

type ServiceConfig struct {
	Server             ServerConfig       `mapstructure:"server"`
	DbConfig           DBConfig           `mapstructure:"database"`
	BrokerConfig       BrokerConfig       `mapstructure:"broker"`
	WorkerConfig       WorkerConfig       `mapstructure:"worker"`
	CloudStorageConfig CloudStorageConfig `mapstructure:"cloud"`
	ProcessingConfig   ProcessingConfig   `mapstructure:"processing"`
}

type DBConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	DBName          string        `yaml:"dbname"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime"`
	DBConn          string
}

type ServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
}

type BrokerConfig struct {
	Brokers           []string      `yaml:"brokers"`
	Topic             string        `yaml:"topic"`
	ConsumerGroup     string        `yaml:"consumerGroup"`
	SessionTimeout    time.Duration `yaml:"sessionTimeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeatInterval"`
	RebalanceTimeout  time.Duration `yaml:"rebalanceTimeout"`
	MaxProcessingTime time.Duration `yaml:"maxProcessingTime"`
	FetchMinBytes     int           `yaml:"fetchMinBytes"`
	FetchMaxBytes     int           `yaml:"fetchMaxBytes"`
	CommitInterval    time.Duration `yaml:"commitInterval"`
}

type WorkerConfig struct {
	NumWorkers        int           `yaml:"numWorkers"`
	BatchSize         int           `yaml:"batchSize"`
	ProcessingTimeout time.Duration `yaml:"processingTimeout"`
	RetryAttempts     int           `yaml:"retryAttempts"`
	RetryBackoff      time.Duration `yaml:"retryBackoff"`
}

type CloudStorageConfig struct {
	Endpoint           string        `yaml:"endpoint"`
	Region             string        `yaml:"region"`
	AccessKey          string        `yaml:"accessKey"`
	SecretKey          string        `yaml:"secretKey"`
	Bucket             string        `yaml:"bucket"`
	UseSSL             bool          `yaml:"useSSL"`
	PresignedURLExpiry time.Duration `yaml:"presignedURLExpiry"`
	UploadPartSize     int64         `yaml:"uploadPartSize"`
	MaxUploadSize      int64         `yaml:"maxUploadSize"`
}

type ProcessingConfig struct {
	DefaultThumbnailSize int      `yaml:"defaultThumbnailSize"`
	DefaultJpegQuality   int      `yaml:"defaultJpegQuality"`
	MaxImageWidth        int      `yaml:"maxImageWidth"`
	MaxImageHeight       int      `yaml:"maxImageHeight"`
	WatermarkOpacity     float64  `yaml:"watermarkOpacity"`
	WatermarkFontSize    int      `yaml:"watermarkFontSize"`
	WatermarkText        string   `yaml:"watermarkText"`
	SupportedFormats     []string `yaml:"supportedFormats"`
}
