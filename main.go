package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"smart-home-backend/devicesCrud"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Could not load env file")
	}
	dbName := os.Getenv("DATABASE_NAME")
	dbUser := os.Getenv("DATABASE_USERNAME")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	connectionStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		log.Fatal("Could not connect to database")
	}

	//////////////////////// HANDLERS //////////////////////////
	// NOTE: DON'T use patch request hangs
	http.HandleFunc("POST /iot-devices", devicesCrud.AddDeviceHandler(db))
	http.HandleFunc("POST /iot-devices/{id}", devicesCrud.EditDeviceHandler(db))
	http.HandleFunc("DELETE /iot-devices/{id}", devicesCrud.DeleteDeviceHandler(db))
	http.HandleFunc("GET /iot-devices", devicesCrud.GetDeviceHandler(db))

	// listen and serv on port 8080
	// uses default standard lib router for
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Could not start server")
	}
	log.Default().Println("Server started")
}
