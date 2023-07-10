package main

import (
	"binanceMS/pkg/db"
	"binanceMS/pkg/trade_cron"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Init() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error getting env")
	} else {
		log.Println("Getting Values")
	}

	db.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

}

func main() {

	Init()
	trade_cron := trade_cron.TradeCron{}
	trade_cron.DB = db.DB

	trade_cron.Run()

	select {}
}
