package models

type Scoreboard struct {
	Data  *ScoreboardData
	Error string
	URL   string
}

type ScoreboardData struct {
	HomeTeam Team
	AwayTeam Team
	Venue    string
	Time     MatchTime
	Status   Status
	League   League
	MatchUrl string
}

type Team struct {
	Name      string
	Url       string
	LogoUrl   string
	Score     string
	IsLeading bool
}

type Status struct {
	Clock       string
	Description string
	IsLive      bool
}

type MatchTime struct {
	Date string
	Time string
}

type EspnApiResponse struct {
	Sports []Sport `json:"sports"`
}

type Sport struct {
	Leagues []League `json:"leagues"`
}

type League struct {
	Name   string  `json:"name"`
	Events []Event `json:"events"`
}

type Event struct {
	Date        string       `json:"date"`
	Location    string       `json:"location"`
	Link        string       `json:"link"`
	FullStatus  FullStatus   `json:"fullStatus"`
	Competitors []Competitor `json:"competitors"`
}

type FullStatus struct {
	DisplayClock string `json:"displayClock"`
	Type         Type   `json:"type"`
}

type Type struct {
	ID          string `json:"id"`
	State       string `json:"state"`
	Description string `json:"description"`
}

type Competitor struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Score       string `json:"score"`
	LogoDark    string `json:"logoDark"`
}
