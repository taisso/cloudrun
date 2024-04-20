package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WeatherApi struct {
	ApiKey string
}

func NewWeather(apiKey string) *WeatherApi {
	return &WeatherApi{ApiKey: apiKey}
}

func (w *WeatherApi) GetWeather(lat, lon string) (*Weather, error) {
	weatherUrl := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s,%s", w.ApiKey, lat, lon)
	weather, err := fetch[Weather](weatherUrl)

	if err != nil {
		return nil, err
	}

	return weather, nil
}

func fetch[T any](url string) (*T, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data T
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
