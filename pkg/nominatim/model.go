package nominatim

type INominatim interface {
	GetLocation(cep string) (*Nominatim, error)
}

type Nominatim struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	Category    string   `json:"category"`
	Type        string   `json:"type"`
	PlaceRank   int      `json:"place_rank"`
	Importance  float64  `json:"importance"`
	Addresstype string   `json:"addresstype"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Boundingbox []string `json:"boundingbox"`
}
