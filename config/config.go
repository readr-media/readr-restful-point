package config

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/viper"
)

var Config AppConfig

type AppConfig struct {
	SQL struct {
		Host                    string                       `mapstructure:"host"`
		Port                    int                          `mapstructure:"port"`
		User                    string                       `mapstructure:"user"`
		Password                string                       `mapstructure:"password"`
		SchemaPath              string                       `mapstructure:"schema_path"`
		TableMeta               map[string]map[string]string `mapstructure:"table_meta"`
		TrasactionIDPlaceholder string                       `mapstructure:"trasaction_id_placeholder"`
	} `mapstructure:"sql"`

	Models struct {
		TaggingType   map[string]int `mapstructure:"tagging_type"`
		FollowingType map[string]int `mapstructure:"following_type"`
		PointType     map[string]int `mapstructure:"point_type"`
		PointStatus   map[string]int `mapstructure:"point_status"`
	} `mapstructure:"models"`

	DomainName string `mapstructure:"domain_name"`

	PaymentService struct {
		PartnerKey         string `mapstructure:"partner_key"`
		MerchantID         string `mapstructure:"merchant_id"`
		PrimeURL           string `mapstructure:"prime_url"`
		TokenURL           string `mapstructure:"token_url"`
		Currency           string `mapstructure:"currency"`
		PaymentDescription string `mapstructure:"payment_description"`
	} `mapstructure:"payment_service"`
}

func LoadConfig(configPath string) (conf AppConfig, err error) {

	viper.SetConfigType("json")
	viper.AutomaticEnv()        // read in environment variables that match
	viper.SetEnvPrefix("READR") // READR_
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if configPath != "" {
		content, err := ioutil.ReadFile(configPath)

		if err != nil {
			return Config, err
		}

		if err := viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return Config, err
		}

		if err := viper.Unmarshal(&Config); err != nil {
			log.Fatalf("Error unmarshal config file, %s", err)
			return Config, err
		}

	} else {
		// Default path
		viper.AddConfigPath("./config/")
		viper.SetConfigName("main")

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file, %s", err)
			return Config, err
		}
		log.Println("Using config file:", viper.ConfigFileUsed())

		if err := viper.Unmarshal(&Config); err != nil {
			log.Fatalf("Error unmarshal config file, %s", err)
			return Config, err
		}
	}

	return Config, nil
}
