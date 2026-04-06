package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DB          DBConfig
	JWTSecret   string
	WorkerCount int
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
			return 0, errors.New("config: wrong parameter for port")
		}
	}
	return port, nil
}

func getDBConfig() (DBConfig, error) {
	config := DBConfig{}
	if env := os.Getenv("DB_HOST"); env == "" {
		return DBConfig{}, errors.New("config: no dbHost info")
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
			return DBConfig{}, errors.New("config: wrong parameter for dbPort")
		}
	} else {
		return DBConfig{}, errors.New("config: no dbPort info")
	}
	if env := os.Getenv("DB_USER"); env == "" {
		return DBConfig{}, errors.New("config: no user info")
	} else {
		config.User = env
	}

	if env := os.Getenv("DB_PASSWORD"); env == "" {
		return DBConfig{}, errors.New("config: no password info")
	} else {
		config.Password = env
	}

	if env := os.Getenv("DB_NAME"); env == "" {
		return DBConfig{}, errors.New("config: no DB_NAME info")
	} else {
		config.Name = env
	}
	return config, nil

}

func getJWTSecret() (string, error) {
	env := os.Getenv("JWT_SECRET")
	if env == "" {
		return "", errors.New("config: no JWT info")
	}
	return env, nil

}

func getWorkerCount() (int, error) {
	env := os.Getenv("WORKER_COUNT")
	if env == "" {
		return 0, errors.New("config: no WORKER_COUNT info")
	}
	envInt, err := strconv.Atoi(env)
	if err != nil {
		return 0, err
	}
	if envInt < 1 {
		return 0, errors.New("config: WORKER_COUNT is less than 1")
	}
	return envInt, nil
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
	workerCount, err := getWorkerCount()
	if err != nil {
		return Config{}, err
	}
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return Config{}, err
	}
	conf := Config{
		Port:        port,
		DB:          confDB,
		JWTSecret:   jwtSecret,
		WorkerCount: workerCount,
	}
	return conf, nil
}
