package config

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Address        string `mapstructure:"RUN_ADDRESS"`
	DatabaseURI    string `mapstructure:"DATABASE_URI"`
	AccrualAddress string `mapstructure:"ACCRUAL_SYSTEM_ADDRESS"`
}

func LoadConfig() (config *Config, err error) {
	address := flag.String("a", "", "The address to listen on for HTTP requests.")
	databaseURI := flag.String("d", ":memory:", "The database URI to use")
	accrualAddress := flag.String("r", "", "The accrual address")
	flag.Parse()

	v := viper.New()

	v.AddConfigPath("./config")
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return
	}
	fmt.Printf("Read config=%v\n", config)
	if *address != "" {
		config.Address = *address
	}
	if *databaseURI != "" {
		config.DatabaseURI = *databaseURI
	}
	if *accrualAddress != "" {
		config.AccrualAddress = *accrualAddress
	}

	return
}
