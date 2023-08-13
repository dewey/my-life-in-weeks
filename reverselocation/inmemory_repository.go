package reverselocation

type InMemoryRepository struct{}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) Address(coordinates string) (Address, error) {
	return Address{
		Latitude:    1.0,
		Longitude:   2.0,
		CountryCode: "DEU",
		Region:      "Berlin",
	}, nil
}
