package config

import (
	"github.com/shipa988/hw_otus_architect/internal/data/controller/queue"
	"time"
)

type Config struct {
	Log      Log    `yaml:"log"`  //if logging to file
	Name     string `yaml:"name"` //app name in logs
	Env      string `yaml:"env"`  //prod or dev
	API      API    `yaml:"api"`
	DB       DB     `yaml:"db"`
	Port     string `yaml:"port"`
	GRPCPort string `yaml:"grpcport"`
	Cache    Cache  `yaml:"cache"`
	Queue    Queue  `yaml:"queue"`
}
type Log struct {
	File string `yaml:"file"`
}

type Cache struct {
	Size int `yaml:"size"`
}

type Queue struct {
	Natsconnection queue.NatsConnection `yaml:"natsconnection"`
	Hub    struct {
		Stanconnection queue.StanConnection `yaml:"stanconnection"`
	} `yaml:"hub"`
	News    struct {
		Stanconnection queue.StanConnection `yaml:"stanconnection"`
	} `yaml:"news"`
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
