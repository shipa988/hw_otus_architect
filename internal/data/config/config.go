package config

import "time"

type Config struct {
	Log                 Log                `yaml:"log"`//if logging to file
	Name                string             `yaml:"name"`//app name in logs
	Env                 string             `yaml:"env"`//prod or dev
	API      API    `yaml:"api"`
	DB       DB     `yaml:"db"`
	Port           string        `yaml:"port"`
}
type Log struct {
	File  string `yaml:"file"`
}

type API struct {
	ReadTimeoutMs  time.Duration `yaml:"readtimeout"`
	WriteTimeoutMs time.Duration `yaml:"writetimeout"`
}

type DB struct {
	Provider string `yaml:"provider"`
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

