package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/taisso/cloudrun/pkg/nominatim"
	"github.com/taisso/cloudrun/pkg/weather"
)

type Response struct {
	TempC float32 `json:"temp_c"`
	TempF float32 `json:"temp_f"`
	TempK float32 `json:"temp_k"`
}

type ResponseError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

var (
	ErrInvalidZipCode       = errors.New("invalid zipcode")
	ErrStatusInternalServer = errors.New("internal error")
	ErrNotFoundZipCode      = errors.New("can not find zipcode")
)

type CepTemperature struct {
	weather   weather.IWeather
	nominatim nominatim.INominatim
}

func NewCepTemperature(weather weather.IWeather, nomination nominatim.INominatim) *CepTemperature {
	return &CepTemperature{weather: weather, nominatim: nomination}
}

func (ct CepTemperature) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ct.GetCepTemperatureHandler(w, r)
}

func (ct CepTemperature) GetCepTemperatureHandler(w http.ResponseWriter, r *http.Request) {
	cep := r.PathValue("cep")

	if len(cep) != 8 {
		ToError(w, http.StatusUnprocessableEntity, ErrInvalidZipCode)
		return
	}

	nominatimResponse, err := ct.nominatim.GetLocation(cep)
	if err != nil {
		if err == nominatim.ErrNotFound {
			ToError(w, http.StatusNotFound, ErrNotFoundZipCode)
			return
		}
		log.Println(err)
		ToError(w, http.StatusInternalServerError, ErrStatusInternalServer)
		return
	}

	lat := nominatimResponse.Lat
	lng := nominatimResponse.Lon

	weatherResponse, err := ct.weather.GetWeather(lat, lng)
	if err != nil {
		log.Println(err)
		ToError(w, http.StatusInternalServerError, ErrStatusInternalServer)
		return
	}

	ToJson(w, http.StatusOK, Response{
		TempC: float32(weatherResponse.Current.TempC),
		TempF: float32(weatherResponse.Current.TempF),
		TempK: float32(weatherResponse.Current.TempC + 273),
	})
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	apiKey := os.Getenv("WEATHER_API_KEY")
	cepTemperature := NewCepTemperature(weather.NewWeather(apiKey), nominatim.NewNominatim())
	http.Handle("/{cep}", cepTemperature)

	http.ListenAndServe(":8080", nil)
}

func ToJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(err)
	}
}

func ToError(w http.ResponseWriter, statusCode int, err error) {
	ToJson(
		w,
		statusCode,
		ResponseError{
			Code:    statusCode,
			Message: err.Error(),
		},
	)
}
