package data

import (
	"dc/internal/config"
	"fmt"

	"github.com/go-redis/redis/v7"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 客户端连接
type Data struct {
	db      *gorm.DB
	redisdb *redis.Client
}

func NewData(conf *config.ConfigData) (*Data, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Database.User,
		conf.Database.PassWord,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Errorf("mysql lost: %v", err))
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.PassWord,
		DB:       0,
	})
	_, err = redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	d := &Data{
		db:      db,
		redisdb: redisClient,
	}
	return d, nil

}
