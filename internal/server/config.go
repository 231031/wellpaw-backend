package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/231031/pethealth-backend/internal/applogger"
	"github.com/231031/pethealth-backend/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt"
)

var (
	serverLog = "[SERVER LOGGER]"
)

type Cfg struct {
	BACKEND_PORT   string
	DB_HOST        string
	DB_PORT        string
	DB_USER        string
	DB_PASSWORD    string
	DB_NAME        string
	REDIS_HOST     string
	REDIS_PORT     string
	REDIS_PASSWORD string
	SECRET_KEY     string
	FILE_PUB_PATH  string
	FILE_PRI_PATH  string
}

func getAllENV() *Cfg {
	allKey := []string{
		"BACKEND_PORT",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"REDIS_HOST",
		"REDIS_PORT",
		"REDIS_PASSWORD",
		"SECRET_KEY",
		"FILE_PUB_PATH",
		"FILE_PRI_PATH",
	}

	allValue := make(map[string]string)
	for _, key := range allKey {
		if os.Getenv(key) == "" {
			panic("Environment variable " + key + " is not set")
		} else {
			allValue[key] = os.Getenv(key)
		}
	}

	cfg := &Cfg{
		BACKEND_PORT:   allValue[allKey[0]],
		DB_HOST:        allValue[allKey[1]],
		DB_PORT:        allValue[allKey[2]],
		DB_USER:        allValue[allKey[3]],
		DB_PASSWORD:    allValue[allKey[4]],
		DB_NAME:        allValue[allKey[5]],
		REDIS_HOST:     allValue[allKey[6]],
		REDIS_PORT:     allValue[allKey[7]],
		REDIS_PASSWORD: allValue[allKey[8]],
		SECRET_KEY:     allValue[allKey[9]],
		FILE_PUB_PATH:  allValue[allKey[10]],
		FILE_PRI_PATH:  allValue[allKey[11]],
	}

	return cfg
}

func InitLogger(app *fiber.App) *os.File {
	file, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		applogger.LogError(fmt.Sprintf("error opening file: %v", err), serverLog)
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	// Set config for logger
	loggerConfig := logger.Config{
		Output:     multiWriter,
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Bangkok",
	}

	app.Use(logger.New(loggerConfig))

	return file
}

func ConfigGenerateKey(cfg *Cfg) *model.TokenConfig {
	tokenCfg := &model.TokenConfig{
		AccessTokenExpirationSecs: 1 * 3600,
		RefreshExpirationSecs:     48 * 3600,
		RefreshSecret:             cfg.SECRET_KEY,
	}

	priv, err := ioutil.ReadFile(cfg.FILE_PRI_PATH)
	if err != nil {
		fmt.Sprintln("failed to read private pem file", err)
		return tokenCfg
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(priv)
	if err != nil {
		fmt.Sprintln("failed to parse private pem to rsa", err)
		return tokenCfg
	}
	tokenCfg.PrivateKey = privateKey

	pub, err := ioutil.ReadFile(cfg.FILE_PUB_PATH)
	if err != nil {
		fmt.Sprintln("failed to read public pem file", err)
		return tokenCfg
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pub)
	if err != nil {
		fmt.Sprintln("failed to parse public pem to rsa", err)
		return tokenCfg
	}
	tokenCfg.PublicKey = publicKey

	return tokenCfg
}
