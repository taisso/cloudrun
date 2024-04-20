package nominatim

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound = errors.New("empty")
)

type NominatimApi struct{}

func NewNominatim() *NominatimApi {
	return &NominatimApi{}
}

func (n *NominatimApi) GetLocation(cep string) (*Nominatim, error) {
	nominatimUrl := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=jsonv2", cep)
	nominatims, err := fetch[[]Nominatim](nominatimUrl)

	if err != nil {
		return nil, err
	}

	if len(*nominatims) == 0 {
		return nil, ErrNotFound
	}

	nominatim := (*nominatims)[0]

	return &nominatim, nil
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
