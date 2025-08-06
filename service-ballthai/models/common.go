package models

// "database/sql" // Removed: 'database/sql' is not directly used in this file

// NationalityAPI represents the nationality data from the API
type NationalityAPI struct {
	Code     string `json:"code"`
	FullName string `json:"full_name"` // API may use "name" or "full_name"
	Name     string `json:"name"`      // Some APIs might use "name" instead of "full_name"
}

// ChannelAPI represents the channel data from the API
type ChannelAPI struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

// CountryAPI represents the country data from the API (used in Stadiums)
type CountryAPI struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// ClubNameAPI represents club name data from the API (used in Stadiums)
type ClubNameAPI struct {
	TH string `json:"th"`
	EN string `json:"en"`
}

// BaseDB represents common fields for database entities (optional, not directly used in all DB structs here)
type BaseDB struct {
	ID int
}
