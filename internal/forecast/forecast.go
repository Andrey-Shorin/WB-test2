package forecast

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/internal/config"
	"net/http"
)

type Root struct {
	Cod     string            `json:"cod"`
	Message int               `json:"message"`
	Cnt     int               `json:"cnt"`
	List    []json.RawMessage `json:"list"`
	City    City              `json:"city"`
}

// City представляет объект "city" в JSON
type City struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Coord      Coord  `json:"coord"`
	Country    string `json:"country"`
	Population int    `json:"population"`
	Timezone   int    `json:"timezone"`
	Sunrise    int64  `json:"sunrise"`
	Sunset     int64  `json:"sunset"`
}

// Coord представляет объект "coord" в JSON
type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type List struct {
	Dt   int64 `json:"dt"`
	Main Main  `json:"main"`

	Visibility int     `json:"visibility"`
	Pop        float64 `json:"pop"`

	DtTxt string `json:"dt_txt"`
}

// Main представляет объект "main" в JSON
type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	SeaLevel  int     `json:"sea_level"`
	GrndLevel int     `json:"grnd_level"`
	Humidity  int     `json:"humidity"`
	TempKf    float64 `json:"temp_kf"`
}

func GetForecast(config config.Config, lat, lon float32) ([]json.RawMessage, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/forecast?lat=" + fmt.Sprintf("%f", lat) + "&lon=" + fmt.Sprintf("%f", lon) + "&appid=" + config.Appid)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var root Root
	err = json.Unmarshal([]byte(body), &root)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return root.List, err

}
func GetWeather(str string) (temp float64, date int64) {
	var list List
	err := json.Unmarshal([]byte(str), &list)
	if err != nil {
		log.Print(err)
		return 0, 0
	}
	return list.Main.Temp, list.Dt
}
