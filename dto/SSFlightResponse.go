package dto

type SSFlightResponse struct {
	Data    Data   `json:"data"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type Data struct {
	Context             Context       `json:"context"`
	Itineraries         []Itinerary   `json:"itineraries"`
	Messages            []interface{} `json:"messages"`
	FilterStats         FilterStats   `json:"filterStats"`
	FlightsSessionID    string        `json:"flightsSessionId"`
	DestinationImageURL string        `json:"destinationImageUrl"`
	Token               string        `json:"token"`
}

type Context struct {
	Status             string `json:"status"`
	SessionID          string `json:"sessionId"`
	TotalResults       int    `json:"totalResults"`
	FilterTotalResults int    `json:"filterTotalResults"`
}

type Itinerary struct {
	ID                      string      `json:"id"`
	Price                   Price       `json:"price"`
	Legs                    []Leg       `json:"legs"`
	IsSelfTransfer          bool        `json:"isSelfTransfer"`
	IsProtectedSelfTransfer bool        `json:"isProtectedSelfTransfer"`
	FarePolicy              FarePolicy  `json:"farePolicy"`
	Eco                     Eco         `json:"eco"`
	FareAttributes          interface{} `json:"fareAttributes"`
	Tags                    []string    `json:"tags"`
	IsMashUp                bool        `json:"isMashUp"`
	HasFlexibleOptions      bool        `json:"hasFlexibleOptions"`
	Score                   float64     `json:"score"`
}

type Price struct {
	Raw             float64 `json:"raw"`
	Formatted       string  `json:"formatted"`
	PricingOptionID string  `json:"pricingOptionId"`
}

type Leg struct {
	ID                string    `json:"id"`
	Origin            Airport   `json:"origin"`
	Destination       Airport   `json:"destination"`
	DurationInMinutes int       `json:"durationInMinutes"`
	StopCount         int       `json:"stopCount"`
	IsSmallestStops   bool      `json:"isSmallestStops"`
	Departure         string    `json:"departure"`
	Arrival           string    `json:"arrival"`
	TimeDeltaInDays   int       `json:"timeDeltaInDays"`
	Carriers          Carriers  `json:"carriers"`
	Segments          []Segment `json:"segments"`
}

type Airport struct {
	ID            string `json:"id"`
	EntityID      string `json:"entityId"`
	Name          string `json:"name"`
	DisplayCode   string `json:"displayCode"`
	City          string `json:"city"`
	Country       string `json:"country"`
	IsHighlighted bool   `json:"isHighlighted"`
}

type Carriers struct {
	Marketing     []Carrier `json:"marketing"`
	OperationType string    `json:"operationType"`
}

type Carrier struct {
	ID          int    `json:"id"`
	AlternateID string `json:"alternateId"`
	LogoURL     string `json:"logoUrl"`
	Name        string `json:"name"`
}

type Segment struct {
	ID                string      `json:"id"`
	Origin            Place       `json:"origin"`
	Destination       Place       `json:"destination"`
	Departure         string      `json:"departure"`
	Arrival           string      `json:"arrival"`
	DurationInMinutes int         `json:"durationInMinutes"`
	FlightNumber      string      `json:"flightNumber"`
	MarketingCarrier  CarrierInfo `json:"marketingCarrier"`
	OperatingCarrier  CarrierInfo `json:"operatingCarrier"`
}

type Place struct {
	FlightPlaceID string `json:"flightPlaceId"`
	DisplayCode   string `json:"displayCode"`
	Parent        Parent `json:"parent"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Country       string `json:"country"`
}

type Parent struct {
	FlightPlaceID string `json:"flightPlaceId"`
	DisplayCode   string `json:"displayCode"`
	Name          string `json:"name"`
	Type          string `json:"type"`
}

type CarrierInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	AlternateID string `json:"alternateId"`
	AllianceID  int    `json:"allianceId"`
	DisplayCode string `json:"displayCode"`
}

type FarePolicy struct {
	IsChangeAllowed       bool `json:"isChangeAllowed"`
	IsPartiallyChangeable bool `json:"isPartiallyChangeable"`
	IsCancellationAllowed bool `json:"isCancellationAllowed"`
	IsPartiallyRefundable bool `json:"isPartiallyRefundable"`
}

type Eco struct {
	EcoContenderDelta float64 `json:"ecoContenderDelta"`
}

type FilterStats struct {
	Duration   Duration       `json:"duration"`
	Airports   []AirportStats `json:"airports"`
	Carriers   []Carrier      `json:"carriers"`
	StopPrices StopPrices     `json:"stopPrices"`
}

type Duration struct {
	Min          int `json:"min"`
	Max          int `json:"max"`
	MultiCityMin int `json:"multiCityMin"`
	MultiCityMax int `json:"multiCityMax"`
}

type AirportStats struct {
	City     string    `json:"city"`
	Airports []Airport `json:"airports"`
}

type StopPrices struct {
	Direct    StopPriceDetail `json:"direct"`
	One       StopPriceDetail `json:"one"`
	TwoOrMore StopPriceDetail `json:"twoOrMore"`
}

type StopPriceDetail struct {
	IsPresent      bool   `json:"isPresent"`
	FormattedPrice string `json:"formattedPrice,omitempty"`
}
