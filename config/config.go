package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	WorkerPool WorkerPool
}

type WorkerPool struct {
	NumWorker          int
	MaxQueueMultiplier int
	SleepDuration      time.Duration
}

func GetConfig() *Config {
	viper.SetConfigFile("config.yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("cannot read config file: ", err.Error())
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("cannot unmarshal config: ", err.Error())
	}

	return &config
}
