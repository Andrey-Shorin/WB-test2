package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DbURL string
	Cytis []string
	Appid string
}

func ReadConfig() Config {
	var conf Config

	env, err := godotenv.Read("conf.env")
	if err != nil {
		log.Fatal("cant read config file")
	}
	conf.DbURL = env["dbURL"]
	conf.Cytis = strings.Split(env["cytis"], ",")
	conf.Appid = env["appid"]

	return conf
}
