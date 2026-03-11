package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Port int
	DB   DBConfig
}
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func getPort() (int, error) {
	port := 7540
	if env := os.Getenv("PORT"); env != "" {
		p, err := strconv.Atoi(env)
		if err != nil {
			return 0, err
		}
		if p > 0 {
			port = p
		} else {
			return 0, errors.New("wrong parameter for port")
		}
	}
	return port, nil
}

func getDBConfig() (DBConfig, error) {
	config := DBConfig{}
	if env := os.Getenv("DB_HOST"); env == "" {
		return DBConfig{}, errors.New("no dbHost info")
	} else {
		config.Host = env
	}

	if env := os.Getenv("DB_PORT"); env != "" {
		p, err := strconv.Atoi(env)
		if err != nil {
			return DBConfig{}, err
		}
		if p > 0 {
			config.Port = p
		} else {
			return DBConfig{}, errors.New("wrong parameter for dbPort")
		}
	} else {
		return DBConfig{}, errors.New("no dbPort info")
	}
	if env := os.Getenv("DB_USER"); env == "" {
		return DBConfig{}, errors.New("no user info")
	} else {
		config.User = env
	}

	if env := os.Getenv("DB_PASSWORD"); env == "" {
		return DBConfig{}, errors.New("no password info")
	} else {
		config.Password = env
	}

	if env := os.Getenv("DB_NAME"); env == "" {
		return DBConfig{}, errors.New("no name info")
	} else {
		config.Name = env
	}

	return config, nil

}

func Load() (Config, error) {
	port, err := getPort()
	if err != nil {
		return Config{}, err
	}
	confDB, err := getDBConfig()
	if err != nil {
		return Config{}, err
	}
	conf := Config{
		Port: port,
		DB:   confDB,
	}
	return conf, nil

}
