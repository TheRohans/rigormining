package models

// Session represents the part of the JWT that has our user id data
// that we care about
type Session struct {
	// Sub is the athleteId - its the same as Username, use username
	Sub      string `json:"sub"`
	EventId  string `json:"event_id"`
	TokenUse string `json:"token_use"`
	Scope    string `json:"scope"`
	AuthTime int32  `json:"auth_time"`
	Iss      string `json:"iss"`
	Exp      int32  `json:"exp"`
	Iat      int32  `json:"iat"`
	Jti      string `json:"jti"`
	// ClientId is the id for conginito, you want Username or Sub
	ClientId string `json:"client_id"`
	// Username will contain the athlete UUID
	Username string `json:"username"`
}

type Researcher struct {
	Id    *string `db:"uuid"`
	Email *string `db:"email"`
	Name  *string `db:"name"`
}

// Highlight is a single highlight
type Highlight struct {
	ResearcherId string `db:"researcher_uuid"`
	Id           string `db:"uuid"`
}
