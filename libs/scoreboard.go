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

func (s *Server) FetchScoreboard(ctx context.Context) Scoreboard {
	url := fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=soccer&team=%s", s.cfg.Scoreboard.Team)
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Scoreboard{Error: fmt.Sprintf("failed to create request: %s", err)}
	}
	rq.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate;")

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
