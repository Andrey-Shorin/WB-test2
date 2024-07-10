package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/internal/config"
	"main/internal/forecast"
	"main/internal/mydb"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type City struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

var dbpool *pgxpool.Pool

func getCityList(w http.ResponseWriter, r *http.Request) {
	rows, err := dbpool.Query(context.Background(), "SELECT id, name, country, lat, lon FROM city ORDER BY name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var city City
		if err := rows.Scan(&city.ID, &city.Name, &city.Country, &city.Lat, &city.Lon); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cities = append(cities, city)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}

type CityForecastSummary struct {
	Name           string   `json:"name"`
	Country        string   `json:"country"`
	AverageTemp    float64  `json:"average_temp"`
	AvailableDates []string `json:"available_dates"`
}

func getCityForecastSummary(w http.ResponseWriter, r *http.Request) {
	cityID := chi.URLParam(r, "id")

	var city City
	err := dbpool.QueryRow(context.Background(), "SELECT id, name, country, lat, lon FROM city WHERE id = $1", cityID).Scan(&city.ID, &city.Name, &city.Country, &city.Lat, &city.Lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := dbpool.Query(context.Background(), "SELECT temp, time FROM forecast WHERE city = $1 AND time > $2 ORDER BY time", city.Name, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var totalTemp float64
	var count int
	var availableDates []string

	for rows.Next() {
		var temp float64
		var timestamp time.Time
		if err := rows.Scan(&temp, &timestamp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		totalTemp += temp
		count++
		availableDates = append(availableDates, fmt.Sprint(timestamp.Unix()))
	}

	if count == 0 {
		http.Error(w, "No forecasts available", http.StatusNotFound)
		return
	}

	averageTemp := totalTemp / float64(count)
	summary := CityForecastSummary{
		Name:           city.Name,
		Country:        city.Country,
		AverageTemp:    averageTemp,
		AvailableDates: availableDates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

type ForecastDetail struct {
	City string          `json:"city"`
	Temp float64         `json:"temp"`
	Time int64           `json:"time"`
	Data json.RawMessage `json:"data"`
}

func getForecastByDate(w http.ResponseWriter, r *http.Request) {
	cityID := chi.URLParam(r, "id")
	d, err := strconv.ParseInt(chi.URLParam(r, "date"), 10, 64)
	if err != nil {
		http.Error(w, "wrong date format", http.StatusNotFound)
	}
	date := time.Unix(d, 0).UTC()

	var city City
	err = dbpool.QueryRow(context.Background(), "SELECT id, name, country, lat, lon FROM city WHERE id = $1", cityID).Scan(&city.ID, &city.Name, &city.Country, &city.Lat, &city.Lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var forecast ForecastDetail
	var T time.Time
	err = dbpool.QueryRow(context.Background(), "SELECT city, temp, time, data FROM forecast WHERE city = $1 AND time::timestamp = $2::timestamp", city.Name, date).Scan(&forecast.City, &forecast.Temp, &T, &forecast.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Forecast not found for the specified date", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	forecast.Time = T.Unix()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forecast)
}

func main() {
	fmt.Println("start")
	conf := config.ReadConfig()
	fmt.Println(conf)
	dbpool = mydb.ConnectDB(conf.DbURL)
	citys := getCitys(conf)
	mydb.WriteCity(citys, dbpool)

	ticker := time.NewTicker(time.Minute)
	done := make(chan bool)
	ubdateForecast(conf, citys)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				ubdateForecast(conf, citys)
			}
		}
	}()

	r := chi.NewRouter()
	r.Get("/cities", getCityList)
	r.Get("/cities/{id:[0-9]+}/forecast", getCityForecastSummary)
	r.Get("/cities/{id:[0-9]+}/forecast/{date:[0-9]+}", getForecastByDate)
	log.Println("Starting server on :8080")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		done <- true
		if err := server.Close(); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}

	}()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
func ubdateForecast(conf config.Config, citys []mydb.MyCity) {

	for _, i := range citys {
		go func() {
			dat, _ := forecast.GetForecast(conf, i.Lat, i.Lon)
			mydb.WriteForecast(dat, i.Name, dbpool)
		}()
	}

}

func getCitys(conf config.Config) []mydb.MyCity {

	var ret []mydb.MyCity
	for _, name := range conf.Cytis {
		resp, err := http.Get("http://api.openweathermap.org/geo/1.0/direct?q=" + name + "&limit=1&appid=" + conf.Appid)
		if err != nil {
			log.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		var responce []mydb.MyCity
		err = json.Unmarshal(body, &responce)
		if err != nil {
			log.Fatal(err)
		}
		ret = append(ret, responce[0])
		resp.Body.Close()
	}
	return ret
}
