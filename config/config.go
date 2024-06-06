package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go-qirania/utils/milog"
)

var Conf Config

type Config struct {
	//Host                      string      `yaml:"Host"`
	BotToken                  string      `yaml:"BotToken"`
	SpreadSheetId             string      `yaml:"SpreadSheetId"`
	TemplateSheetId           int64       `yaml:"TemplateSheetId"`
	CellRange                 string      `yaml:"CellRange"`
	CredentialPath            string      `yaml:"CredentialPath"`
	TokenPath                 string      `yaml:"TokenPath"`
	DelayWhenNoJobInSeconds   int         `yaml:"DelayWhenNoJobInSeconds"`
	DelayWhenErrorInSeconds   int         `yaml:"DelayWhenErrorInSeconds"`
	DelayWhenJobDoneInSeconds int         `yaml:"DelayWhenJobDoneInSeconds"`
	Redis                     RedisConfig `yaml:"Redis"`
}

type RedisConfig struct {
	Host      []string `yaml:"Host"`
	Password  string   `yaml:"Password"`
	DB        int      `yaml:"DB"`
	MaxIdle   int      `yaml:"MaxIdle"`
	MaxActive int      `yaml:"MaxActive"`
}

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		milog.Fatalf("Error reading config file, %s", err)
	}
	err = viper.Unmarshal(&Conf)
	if err != nil {
		milog.Fatalf("unable to decode into struct, %v", err)
	}
	milog.Info("success reading config file")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		err = viper.Unmarshal(&Conf)
		if err != nil {
			milog.Fatalf("unable to decode into struct, %v", err)
		}
		milog.Info("success re-read config file")
	})
}
