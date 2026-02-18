package main

type JieGuo struct {
	IP          string  `json:"ip"`
	Hostname    string  `json:"hostname,omitempty"`
	City        string  `json:"city,omitempty"`
	Region      string  `json:"region,omitempty"`
	RegionCode  string  `json:"region_code,omitempty"`
	Country     string  `json:"country,omitempty"`
	CountryName string  `json:"country_name,omitempty"`
	Continent   string  `json:"continent,omitempty"`
	Postal      string  `json:"postal,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
	UTCOffset   string  `json:"utc_offset,omitempty"`
	Org         string  `json:"org,omitempty"`
	ASN         string  `json:"asn,omitempty"`
	ISP         string  `json:"isp,omitempty"`
	IsProxy     *bool   `json:"is_proxy,omitempty"`
	Sources     int     `json:"_sources"`
}

type resp1 struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

type resp2 struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	RegionCode  string  `json:"region_code"`
	Country     string  `json:"country"`
	CountryName string  `json:"country_name"`
	Postal      string  `json:"postal"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	UTCOffset   string  `json:"utc_offset"`
	Org         string  `json:"org"`
	ASN         string  `json:"asn"`
	Error       bool    `json:"error"`
	Reason      string  `json:"reason"`
}

type resp3 struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	IP          string  `json:"ip"`
	Continent   string  `json:"continent"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	Timezone    string  `json:"timezone"`
	UTCOffset   string  `json:"timezone_gmt"`
}

type resp4 struct {
	IPAddress   string  `json:"ipAddress"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CountryName string  `json:"countryName"`
	CountryCode string  `json:"countryCode"`
	TimeZone    string  `json:"timeZone"`
	ZipCode     string  `json:"zipCode"`
	CityName    string  `json:"cityName"`
	RegionName  string  `json:"regionName"`
	Continent   string  `json:"continent"`
	IsProxy     bool    `json:"isProxy"`
}

type resp5 struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionName  string  `json:"region_name"`
	CityName    string  `json:"city_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ZipCode     string  `json:"zip_code"`
	TimeZone    string  `json:"time_zone"`
	ASN         string  `json:"asn"`
	AS          string  `json:"as"`
	IsProxy     bool    `json:"is_proxy"`
}

type resp6 struct {
	IPAddress   string `json:"ipAddress"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
	StateProv   string `json:"stateProv"`
	City        string `json:"city"`
}

type cacheItem struct {
	data     *JieGuo
	expireAt int64
}
