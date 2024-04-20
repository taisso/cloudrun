package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/taisso/cloudrun/pkg/nominatim"
	"github.com/taisso/cloudrun/pkg/weather"
)

type MockWeatherApi struct {
	mock.Mock
}

func (w *MockWeatherApi) GetWeather(lat, lon string) (*weather.Weather, error) {
	args := w.Called(lat, lon)
	return args.Get(0).(*weather.Weather), args.Error(1)
}

type MockNominatimApi struct {
	mock.Mock
}

func (n *MockNominatimApi) GetLocation(cep string) (*nominatim.Nominatim, error) {
	args := n.Called(cep)
	return args.Get(0).(*nominatim.Nominatim), args.Error(1)
}

func Setup(t *testing.T, param string, cepTemperature *CepTemperature) (*httptest.ResponseRecorder, []byte) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	req.SetPathValue("cep", param)

	cepTemperature.GetCepTemperatureHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	assert.Nil(t, err)

	return w, data
}

func TestZipCodeInvalid(t *testing.T) {
	weatherApi := &MockWeatherApi{}
	nominatimApi := &MockNominatimApi{}

	cepTemperature := NewCepTemperature(weatherApi, nominatimApi)
	w, data := Setup(t, "180862", cepTemperature)

	var response ResponseError
	err := json.Unmarshal(data, &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, ErrInvalidZipCode.Error(), response.Message)
	nominatimApi.AssertNumberOfCalls(t, "GetLocation", 0)
	weatherApi.AssertNumberOfCalls(t, "GetWeather", 0)
}

func TestZipCodeNotFound(t *testing.T) {
	weatherApi := &MockWeatherApi{}
	nominatimApi := &MockNominatimApi{}

	cep := "00000000"

	nominatimApi.On("GetLocation", cep).Return(&nominatim.Nominatim{}, nominatim.ErrNotFound)
	cepTemperature := NewCepTemperature(weatherApi, nominatimApi)
	w, data := Setup(t, cep, cepTemperature)

	var response ResponseError
	err := json.Unmarshal(data, &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, ErrNotFoundZipCode.Error(), response.Message)

	nominatimApi.AssertExpectations(t)
	weatherApi.AssertNumberOfCalls(t, "GetWeather", 0)
}

func TestInternalServer(t *testing.T) {
	weatherApi := &MockWeatherApi{}
	nominatimApi := &MockNominatimApi{}

	cep := "00000000"

	nominatimApi.On("GetLocation", cep).Return(&nominatim.Nominatim{}, ErrStatusInternalServer)
	cepTemperature := NewCepTemperature(weatherApi, nominatimApi)
	w, data := Setup(t, cep, cepTemperature)

	var response ResponseError
	err := json.Unmarshal(data, &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, ErrStatusInternalServer.Error(), response.Message)

	nominatimApi.AssertExpectations(t)
	weatherApi.AssertNumberOfCalls(t, "GetWeather", 0)

	// CASE 2
	weatherApi = &MockWeatherApi{}
	nominatimApi = &MockNominatimApi{}

	cep = "11111111"

	weatherApi.On("GetWeather", "80", "100").Return(&weather.Weather{}, ErrStatusInternalServer)
	nominatimApi.On("GetLocation", cep).Return(&nominatim.Nominatim{PlaceID: 000, Lat: "80", Lon: "100"}, nil)

	cepTemperature = NewCepTemperature(weatherApi, nominatimApi)
	w, data = Setup(t, cep, cepTemperature)

	err = json.Unmarshal(data, &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, ErrStatusInternalServer.Error(), response.Message)

	nominatimApi.AssertExpectations(t)
	weatherApi.AssertExpectations(t)

}

func TestResultSuccess(t *testing.T) {
	weatherApi := &MockWeatherApi{}
	nominatimApi := &MockNominatimApi{}

	cep := "00000000"

	wetherResponse := &weather.Weather{
		Current: weather.Current{TempC: 20.0, TempF: 40},
	}

	weatherApi.On("GetWeather", "80", "100").Return(wetherResponse, nil)
	nominatimApi.On("GetLocation", cep).Return(&nominatim.Nominatim{PlaceID: 000, Lat: "80", Lon: "100"}, nil)
	cepTemperature := NewCepTemperature(weatherApi, nominatimApi)
	w, data := Setup(t, cep, cepTemperature)

	var response Response
	err := json.Unmarshal(data, &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, response.TempC)
	assert.Equal(t, float32(wetherResponse.Current.TempC), response.TempC)
	assert.Equal(t, float32(wetherResponse.Current.TempF), response.TempF)
	assert.NotEmpty(t, response.TempF)
	assert.NotEmpty(t, response.TempK)

	nominatimApi.AssertExpectations(t)
	weatherApi.AssertExpectations(t)
}
