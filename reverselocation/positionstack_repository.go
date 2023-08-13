package reverselocation

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PositionstackRepository struct {
	token string
}

func NewRepository(token string) *PositionstackRepository {
	return &PositionstackRepository{
		token: token,
	}
}

// Address returns the first result for a given coordinate string from the Positionstack API
func (r *PositionstackRepository) Address(coordinates string) (Address, error) {
	resp, err := http.Get(fmt.Sprintf("http://api.positionstack.com/v1/reverse?access_key=%s&query=%s", r.token, coordinates))
	if err != nil {
		return Address{}, err
	}
	//b, _ := io.ReadAll(resp.Body)
	//fmt.Println(string(b))
	if resp.StatusCode != http.StatusOK {
		return Address{}, fmt.Errorf("unexpected status code: %d, wanted: %d", resp.StatusCode, http.StatusOK)
	}
	var psr PositionstackResponse
	if err := json.NewDecoder(resp.Body).Decode(&psr); err != nil {
		return Address{}, err
	}
	if len(psr.Data) > 1 {
		return Address(psr.Data[0]), nil
	}
	//if psr.Data.CountryCode != "" || psr.Data.Region != "" {
	//	return psr.Data, nil
	//}
	return Address{}, fmt.Errorf("no results for coordinates: %s", coordinates)
}

type PositionstackResult struct {
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

type PositionstackResponse struct {
	Data []PositionstackResult `json:"data"`
}

type T struct {
	Data []struct {
		Latitude           float64     `json:"latitude"`
		Longitude          float64     `json:"longitude"`
		Type               string      `json:"type"`
		Distance           float64     `json:"distance"`
		Name               string      `json:"name"`
		Number             interface{} `json:"number"`
		PostalCode         interface{} `json:"postal_code"`
		Street             interface{} `json:"street"`
		Confidence         float64     `json:"confidence"`
		Region             interface{} `json:"region"`
		RegionCode         interface{} `json:"region_code"`
		County             interface{} `json:"county"`
		Locality           interface{} `json:"locality"`
		AdministrativeArea interface{} `json:"administrative_area"`
		Neighbourhood      interface{} `json:"neighbourhood"`
		Country            interface{} `json:"country"`
		CountryCode        interface{} `json:"country_code"`
		Continent          interface{} `json:"continent"`
		Label              string      `json:"label"`
	} `json:"data"`
}
