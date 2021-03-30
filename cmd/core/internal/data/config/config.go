package config

import (
	"github.com/shipa988/hw_otus_architect/internal/data/controller/queue"
	"time"
)

type Config struct {
	Log         Log         `yaml:"log"`  //if logging to file
	Name        string      `yaml:"name"` //app name in logs
	Env         string      `yaml:"env"`  //prod or dev
	API         API         `yaml:"api"`
	DB          DB          `yaml:"db"`
	Port        string      `yaml:"port"`
	NewsService NewsService `yaml:"newsservice"`
}
type Log struct {
	File string `yaml:"file"`
}
type NewsService struct {
	Address string `yaml:"address"`
	Queue   struct {
		Natsconnection queue.NatsConnection `yaml:"natsconnection"`
		Hub            struct {
			Stanconnection queue.StanConnection `yaml:"stanconnection"`
		} `yaml:"hub"`
	}
}

type API struct {
	ReadTimeoutMs  time.Duration `yaml:"readtimeout"`
	WriteTimeoutMs time.Duration `yaml:"writetimeout"`
}

type DB struct {
	Provider string   `yaml:"provider"`
	Master   string   `yaml:"master"`
	Slaves   []string `yaml:"slaves"`
	Login    string   `yaml:"login"`
	Password string   `yaml:"password"`
	Name     string   `yaml:"name"`
}
