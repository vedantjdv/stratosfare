package dto

type FlightRequest struct {
	FromEntityID                string
	ToEntityID                  string
	DepartDate                  string
	Market                      string
	Currency                    string
	Stops                       string
	IncludeOriginNearbyAirports string
	Sort                        string
}
