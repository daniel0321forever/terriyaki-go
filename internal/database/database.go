package database

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Db *gorm.DB //created outside to make it global.

func Connect() (*gorm.DB, error) {
	fmt.Println("Connecting to database...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error is occurred  on .env file please check")
	}

	host := os.Getenv("POSTGRES_HOST")
	port, _ := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	user := os.Getenv("POSTGRES_USER")
	dbname := os.Getenv("POSTGRES_DB")
	pass := os.Getenv("POSTGRES_PASSWORD")
	sslmode := os.Getenv("POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable" // default for local development
	}

	psqlSetup := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		host, port, user, dbname, pass, sslmode)
	fmt.Println("psqlSetup: ", psqlSetup)

	config := &gorm.Config{}
	if os.Getenv("GORM_SILENT") == "true" {
		config.Logger = logger.Default.LogMode(logger.Silent)
	} else {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(psqlSetup), config)

	if err != nil {
		fmt.Println("There is an error while connecting to the database ", err)
		panic(err)
	} else {
		Db = db
		fmt.Println("Successfully connected to database!")
	}

	return db, nil
}

func ConnectTestDB() (*gorm.DB, error) {
	fmt.Println("Connecting to test database...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error is occurred  on .env file please check")
	}

	host := os.Getenv("TEST_POSTGRES_HOST")
	port, _ := strconv.Atoi(os.Getenv("TEST_POSTGRES_PORT"))
	user := os.Getenv("TEST_POSTGRES_USER")
	dbname := os.Getenv("TEST_POSTGRES_DB")
	pass := os.Getenv("TEST_POSTGRES_PASSWORD")
	sslmode := os.Getenv("TEST_POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable" // default for local development
	}

	psqlSetup := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		host, port, user, dbname, pass, sslmode)
	fmt.Println("psqlSetup: ", psqlSetup)

	config := &gorm.Config{}
	if os.Getenv("GORM_SILENT") == "true" {
		fmt.Println("GORM_SILENT is true, setting logger to silent")
		config.Logger = logger.Default.LogMode(logger.Silent)
	} else {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(psqlSetup), config)
	if err != nil {
		fmt.Println("There is an error while connecting to the database ", err)
		panic(err)
	} else {
		Db = db
		fmt.Println("Successfully connected to test database!")
	}

	return db, nil
}
