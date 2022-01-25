package main

import (
	"fmt"

	"github.com/go-redis/redis"
	gorma "github.com/hhxsv5/go-redis-memory-analysis"
)

func main() {
	writeRedis(10000, "len10_10k", generateValue(10))
	writeRedis(50000, "len10_50k", generateValue(10))
	// writeRedis(500000, "len10_500k", generateValue(10))

	// writeRedis(10000, "len1000_10k", generateValue(1000))
	// writeRedis(50000, "len1000_50k", generateValue(1000))
	// writeRedis(500000, "len1000_500k", generateValue(1000))

	// writeRedis(10000, "len5000_10k", generateValue(5000))
	// writeRedis(50000, "len5000_50k", generateValue(5000))
	// writeRedis(500000, "len5000_500k", generateValue(5000))

	redisAnalysis()

}

const (
	redisAdd  string = "10.226.20.22"
	redisPort uint16 = 6379
)

var client *redis.Client

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", redisAdd, redisPort),
		Password: "",
		DB:       0,
	})

}

func writeRedis(num int, key, value string) {
	for i := 0; i < num; i++ {
		redisKey := fmt.Sprintf("%s:%v", key, i)
		err := client.Set(redisKey, value, -1).Err()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func redisAnalysis() {
	analysis, err := gorma.NewAnalysisConnection(redisAdd, redisPort, "")
	if err != nil {
		fmt.Println("something wrong:", err)
		return
	}
	defer analysis.Close()

	analysis.Start([]string{":"})

	err = analysis.SaveReports("./reports")
	if err == nil {
		fmt.Println("done")
	} else {
		fmt.Println("error:", err)
	}
}

func generateValue(size int) string {
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		arr[i] = 'a'
	}
	return string(arr)
}
