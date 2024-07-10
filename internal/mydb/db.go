package mydb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/internal/forecast"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MyCity struct {
	Name    string  `json:"name"`
	Country string  `json:"country"`
	Lat     float32 `json:"lat"`
	Lon     float32 `json:"lon"`
}

func ConnectDB(dbURL string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)

	}

	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return dbpool

}
func WriteCity(cytis []MyCity, dbpool *pgxpool.Pool) (err error) {
	Tx, err := dbpool.Begin(context.Background())
	defer func() {
		if err == nil {
			err = Tx.Commit(context.Background())
		} else {
			Tx.Rollback(context.Background())
		}
	}()

	for _, i := range cytis {

		_, err = Tx.Exec(context.Background(), "INSERT INTO city (name,country, lat,lon ) VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO NOTHING;", i.Name, i.Country, i.Lat, i.Lon)
		if err != nil {
			fmt.Print(err)
			return err
		}

	}
	return nil
}

func WriteForecast(Forecast []json.RawMessage, city string, dbpool *pgxpool.Pool) (err error) {
	Tx, err := dbpool.Begin(context.Background())
	defer func() {
		if err == nil {
			err = Tx.Commit(context.Background())
		} else {
			Tx.Rollback(context.Background())
		}
	}()

	for _, i := range Forecast {
		temp, date := forecast.GetWeather(string(i))
		_, err = Tx.Exec(context.Background(), `
        INSERT INTO forecast (city, temp, time, data)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (city, time) 
        DO UPDATE SET temp = EXCLUDED.temp, data = EXCLUDED.data;
    `, city, temp, time.Unix(date, 0).UTC(), string(i))
		if err != nil {
			fmt.Println(err)
			return err
		}

	}
	return nil
}
