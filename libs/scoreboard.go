package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bachtran02/bachtran.go/models"

	"golang.org/x/exp/slog"
)

func (s *Server) FetchScoreboard(ctx context.Context) models.Scoreboard {
	url := fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=soccer&team=%s", s.cfg.Scoreboard.Team)
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	rq.Header.Set("Cache-Control", "no-cache")
	if err != nil {
		return models.Scoreboard{Error: fmt.Sprintf("failed to create request: %s", err)}
	}

	rs, err := s.httpClient.Do(rq)
	if err != nil {
		return models.Scoreboard{Error: fmt.Sprintf("failed to do request: %s", err)}
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		return models.Scoreboard{Error: fmt.Sprintf("non-OK HTTP status: %d\tReason: %s", rs.StatusCode, http.StatusText(rs.StatusCode))}
	}

	var resp models.EspnApiResponse
	if err = json.NewDecoder(rs.Body).Decode(&resp); err != nil {
		return models.Scoreboard{Error: fmt.Sprintf("failed to decode response: %s", err)}
	}

	var league models.League = resp.Sports[0].Leagues[0]
	var event models.Event = league.Events[0]

	userTime, timeErr := time.Parse(time.RFC3339, event.Date)
	if timeErr != nil {
		return models.Scoreboard{Error: fmt.Sprintf("failed to parse time: %s", err)}
	}
	tz, err := time.LoadLocation(s.cfg.Scoreboard.Timezone)
	if err != nil {
		slog.Error("error loading config timezone:", err)
	} else {
		userTime = userTime.In(tz)
	}

	return models.Scoreboard{
		Data: &models.ScoreboardData{
			AwayTeam: models.Team{
				Name:      event.Competitors[0].DisplayName,
				Url:       fmt.Sprintf("https://www.espn.com/soccer/team/_/id/%s", event.Competitors[0].ID),
				LogoUrl:   event.Competitors[0].LogoDark,
				Score:     event.Competitors[0].Score,
				IsLeading: event.Competitors[0].Score > event.Competitors[1].Score,
			},
			HomeTeam: models.Team{
				Name:      event.Competitors[1].DisplayName,
				Url:       fmt.Sprintf("https://www.espn.com/soccer/team/_/id/%s", event.Competitors[1].ID),
				LogoUrl:   event.Competitors[1].LogoDark,
				Score:     event.Competitors[1].Score,
				IsLeading: event.Competitors[1].Score > event.Competitors[0].Score,
			},
			Venue: event.Location,
			Time: models.MatchTime{
				Date: fmt.Sprintf("%s %d", userTime.Month().String(), userTime.Day()),
				Time: fmt.Sprintf("%02d:%02d", userTime.Hour(), userTime.Minute()),
			},
			Status: models.Status{
				Clock:       getClock(event.FullStatus.Type.ID, event.FullStatus.Type.State, event.FullStatus.DisplayClock, event.FullStatus.Type.Description),
				Description: getDescription(event.FullStatus.Type.State, event.FullStatus.Type.Description),
				IsLive:      event.FullStatus.Type.State == "in",
			},
			League:   league,
			MatchUrl: event.Link,
		},
		URL: url,
	}
}

func getClock(state_id string, state string, clock string, description string) string {
	if state == "in" {
		if state_id == "23" {
			return description
		}
		return clock
	}
	return ""
}

func getDescription(state string, description string) string {
	switch state {
	case "in":
		return description
	case "pre":
		return "Scheduled"
	case "post":
		return "Finished"
	}
	return "unknown_status"
}
