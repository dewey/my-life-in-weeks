package reverselocation

type ReverseLocation interface {
	Address(coordinates string) (Address, error)
}

type Address struct {
	Latitude           float64     `json:"latitude"`
	Longitude          float64     `json:"longitude"`
	Label              string      `json:"label"`
	Name               string      `json:"name"`
	Type               string      `json:"type"`
	Distance           float64     `json:"distance"`
	Number             string      `json:"number"`
	Street             string      `json:"street"`
	PostalCode         string      `json:"postal_code"`
	Confidence         float64     `json:"confidence"`
	Region             string      `json:"region"`
	RegionCode         string      `json:"region_code"`
	AdministrativeArea interface{} `json:"administrative_area"`
	Neighbourhood      string      `json:"neighbourhood"`
	Country            string      `json:"country"`
	CountryCode        string      `json:"country_code"`
	MapUrl             string      `json:"map_url"`
}
