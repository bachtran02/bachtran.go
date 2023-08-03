package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

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

// sponsored by ChatGPT
type EspnApiResponse struct {
	Sports []Sport `json:"sports"`
}

type Sport struct {
	ID      string   `json:"id"`
	UID     string   `json:"uid"`
	GUID    string   `json:"guid"`
	Name    string   `json:"name"`
	Slug    string   `json:"slug"`
	Logos   []Logo   `json:"logos"`
	Leagues []League `json:"leagues"`
}

type Logo struct {
	Href   string   `json:"href"`
	Alt    string   `json:"alt"`
	Rel    []string `json:"rel"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
}

type League struct {
	ID           string   `json:"id"`
	UID          string   `json:"uid"`
	Name         string   `json:"name"`
	Abbreviation string   `json:"abbreviation"`
	ShortName    string   `json:"shortName"`
	Slug         string   `json:"slug"`
	Tag          string   `json:"tag"`
	IsTournament bool     `json:"isTournament"`
	SmartDates   []string `json:"smartdates"`
	Events       []Event  `json:"events"`
}

type Event struct {
	ID                  string       `json:"id"`
	UID                 string       `json:"uid"`
	Date                string       `json:"date"`
	TimeValid           bool         `json:"timeValid"`
	Recent              bool         `json:"recent"`
	Name                string       `json:"name"`
	ShortName           string       `json:"shortName"`
	Links               []Link       `json:"links"`
	GamecastAvailable   bool         `json:"gamecastAvailable"`
	PlayByPlayAvailable bool         `json:"playByPlayAvailable"`
	CommentaryAvailable bool         `json:"commentaryAvailable"`
	OnWatch             bool         `json:"onWatch"`
	CompetitionID       string       `json:"competitionId"`
	Location            string       `json:"location"`
	Season              int          `json:"season"`
	SeasonStartDate     string       `json:"seasonStartDate"`
	SeasonEndDate       string       `json:"seasonEndDate"`
	SeasonType          string       `json:"seasonType"`
	SeasonTypeHasGroups bool         `json:"seasonTypeHasGroups"`
	Group               Group        `json:"group"`
	Link                string       `json:"link"`
	Status              string       `json:"status"`
	Summary             string       `json:"summary"`
	Period              int          `json:"period"`
	Clock               string       `json:"clock"`
	AddedClock          float32      `json:"addedClock"`
	FullStatus          FullStatus   `json:"fullStatus"`
	Competitors         []Competitor `json:"competitors"`
	AppLinks            []AppLink    `json:"appLinks"`
}

type Link struct {
	Rel  []string `json:"rel"`
	Href string   `json:"href"`
	Text string   `json:"text"`
}

type Group struct {
	GroupID      string `json:"groupId"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	ShortName    string `json:"shortName"`
}

type FullStatus struct {
	Clock        float32 `json:"clock"`
	AddedClock   float32 `json:"addedClock"`
	DisplayClock string  `json:"displayClock"`
	Period       int     `json:"period"`
	Type         Type    `json:"type"`
}

type Type struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
	Detail      string `json:"detail"`
	ShortDetail string `json:"shortDetail"`
}

type Competitor struct {
	ID             string `json:"id"`
	UID            string `json:"uid"`
	Type           string `json:"type"`
	Order          int    `json:"order"`
	HomeAway       string `json:"homeAway"`
	Winner         bool   `json:"winner"`
	Form           string `json:"form"`
	DisplayName    string `json:"displayName"`
	Name           string `json:"name"`
	Abbreviation   string `json:"abbreviation"`
	Location       string `json:"location"`
	Color          string `json:"color"`
	AlternateColor string `json:"alternateColor"`
	Score          string `json:"score"`
	IsNational     bool   `json:"isNational"`
	Record         string `json:"record"`
	Logo           string `json:"logo"`
	LogoDark       string `json:"logoDark"`
}

type AppLink struct {
	Rel       []string `json:"rel"`
	Href      string   `json:"href"`
	Text      string   `json:"text"`
	ShortText string   `json:"shortText"`
}

func (s *Server) FetchScoreboard(ctx context.Context) Scoreboard {
	url := fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=soccer&team=%s", s.cfg.Scoreboard.Team)
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Scoreboard{Error: fmt.Sprintf("failed to create request: %s", err)}
	}
	rs, err := s.httpClient.Do(rq)
	if err != nil {
		return Scoreboard{Error: fmt.Sprintf("failed to do request: %s", err)}
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		return Scoreboard{Error: fmt.Sprintf("non-OK HTTP status: %d\tReason: %s", rs.StatusCode, http.StatusText(rs.StatusCode))}
	}

	var resp EspnApiResponse
	if err = json.NewDecoder(rs.Body).Decode(&resp); err != nil {
		return Scoreboard{Error: fmt.Sprintf("failed to decode response: %s", err)}
	}

	var league League = resp.Sports[0].Leagues[0]
	var event Event = league.Events[0]

	userTime, timeErr := time.Parse(time.RFC3339, event.Date)
	if timeErr != nil {
		return Scoreboard{Error: fmt.Sprintf("failed to parse time: %s", err)}
	}
	tz, err := time.LoadLocation(s.cfg.Scoreboard.Timezone)
	if err != nil {
		slog.Error("error loading config timezone:", err)
	} else {
		userTime = userTime.In(tz)
	}

	return Scoreboard{
		Data: &ScoreboardData{
			AwayTeam: Team{
				Name:      event.Competitors[0].DisplayName,
				Url:       fmt.Sprintf("https://www.espn.com/soccer/team/_/id/%s", event.Competitors[0].ID),
				LogoUrl:   event.Competitors[0].LogoDark,
				Score:     event.Competitors[0].Score,
				IsLeading: event.Competitors[0].Score > event.Competitors[1].Score,
			},
			HomeTeam: Team{
				Name:      event.Competitors[1].DisplayName,
				Url:       fmt.Sprintf("https://www.espn.com/soccer/team/_/id/%s", event.Competitors[1].ID),
				LogoUrl:   event.Competitors[1].LogoDark,
				Score:     event.Competitors[1].Score,
				IsLeading: event.Competitors[1].Score > event.Competitors[0].Score,
			},
			Venue: event.Location,
			Time: MatchTime{
				Date: fmt.Sprintf("%s %d", userTime.Month().String(), userTime.Day()),
				Time: fmt.Sprintf("%02d:%02d", userTime.Hour(), userTime.Minute()),
			},
			Status: Status{
				Clock:       getClock(event.FullStatus.Type.State, event.FullStatus.DisplayClock),
				Description: getDescription(event.FullStatus.Type.State, event.FullStatus.Type.Description),
				IsLive:      event.FullStatus.Type.State == "in",
			},
			League:   league,
			MatchUrl: event.Link,
		},
		URL: url,
	}
}

func getClock(state string, clock string) string {
	if state == "in" {
		return clock
	}
	return ""
}

func getDescription(state string, description string) string {
	if state == "in" {
		return description
	} else if state == "pre" {
		return "Scheduled"
	} else {
		return "Finished"
	}
}
