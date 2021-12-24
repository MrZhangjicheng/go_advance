package conf

import (
	"dc/internal/data"
	"fmt"
	"os"

	"dc/internal/config"

	"github.com/joho/godotenv"
)

var SourceDB *data.Data

func Init() {
	str, _ := os.Getwd()
	fmt.Println(str)
	_ = godotenv.Load("../../configs/.env")
	fmt.Println(os.Getenv("KAE_MYSQL_PASSWORD"))
	db := &config.ConfigDB{
		PassWord: os.Getenv("KAE_MYSQL_PASSWORD"),
		User:     os.Getenv("KAE_MYSQL_USER"),
		Host:     os.Getenv("KAE_MYSQL_HOST"),
		Port:     os.Getenv("KAE_MYSQL_PORT"),
		DBName:   os.Getenv("KAE_MYSQL_DB_NAME"),
	}

	rdb := &config.ConfigRDB{
		Addr:     os.Getenv("KAE_REDIS_ADDR"),
		PassWord: os.Getenv("KAE_REDIS_PW"),
	}

	sorcedata := &config.ConfigData{
		Database: db,
		Redis:    rdb,
	}
	var err error
	SourceDB, err = data.NewData(sorcedata)
	if err != nil {
		panic("source load failed")
	}
}
