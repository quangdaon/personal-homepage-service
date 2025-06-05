package ups

type OAuthResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type Status struct {
	Code                  string `json:"code"`
	Description           string `json:"description"`
	SimplifiedDescription string `json:"simplifiedTextDescription"`
	StatusCode            string `json:"statusCode"`
	Type                  string `json:"type"`
}

type Address struct {
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	AddressLine3 string `json:"addressLine3"`
	City         string `json:"city"`
	State        string `json:"stateProvince"`
	PostalCode   string `json:"postalCode"`
	Country      string `json:"country"`
	CountryCode  string `json:"countryCode"`
}

type Location struct {
	Address Address `json:"address"`
}

type Activity struct {
	Location       Location `json:"location"`
	Date           string   `json:"gmtDate"`
	Time           string   `json:"gmtTime"`
	TimeZoneOffset string   `json:"gmtOffset"`
	Status         Status   `json:"status"`
}

type DeliveryDate struct {
	Date string `json:"date"`
	Type string `json:"type"`
}

type DeliveryTime struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Type      string `json:"type"`
}

type Package struct {
	TrackingNumber string         `json:"trackingNumber"`
	DeliveryTime   DeliveryTime   `json:"deliveryTime"`
	DeliveryDate   []DeliveryDate `json:"deliveryDate"`
	CurrentStatus  Status         `json:"currentStatus"`
	Activity       []Activity     `json:"activity"`
}

type Shipment struct {
	InquiryNumber string    `json:"inquiryNumber"`
	Packages      []Package `json:"package"`
}

type TrackingResponse struct {
	Shipments []Shipment `json:"shipment"`
}

type ApiResponse struct {
	Response TrackingResponse `json:"trackResponse"`
}
