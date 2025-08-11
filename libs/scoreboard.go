package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bachtran02/bachtran.go/models"

	"golang.org/x/exp/slog"
)

func (s *Server) doESPNRequest(ctx context.Context, rawURL string) (*http.Response, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	rq.Header.Set("Cache-Control", "no-cache")

	rs, err := s.httpClient.Do(rq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if rs.StatusCode != http.StatusOK {
		defer rs.Body.Close()
		return nil, fmt.Errorf("non-OK HTTP status: %d\tReason: %s", rs.StatusCode, http.StatusText(rs.StatusCode))
	}
	return rs, nil
}

func (s *Server) FetchScoreboard(ctx context.Context) (models.Scoreboard, error) {
	rawURL := fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=soccer&team=%s", s.cfg.Scoreboard.Team)
	rs, err := s.doESPNRequest(ctx, rawURL)
	if err != nil {
		return models.Scoreboard{}, fmt.Errorf("failed to fetch scoreboard: %w", err)
	}
	defer rs.Body.Close()

	var resp models.EspnApiResponse
	if err = json.NewDecoder(rs.Body).Decode(&resp); err != nil {
		return models.Scoreboard{}, fmt.Errorf("failed to decode scoreboard response: %w", err)
	}

	var league models.League = resp.Sports[0].Leagues[0]
	var event models.Event = league.Events[0]

	userTime, timeErr := time.Parse(time.RFC3339, event.Date)
	if timeErr != nil {
		return models.Scoreboard{}, fmt.Errorf("failed to parse time: %w", timeErr)
	}
	tz, err := time.LoadLocation("America/Los_Angeles") /* default server-side rendered timezone */
	if err != nil {
		slog.Error("error loading config timezone:", err)
	} else {
		userTime = userTime.In(tz)
	}

	/* Fetch the static league icon */
	var leagueIconSuffix string
	switch league.Code {
	case "uefa.champions":
		leagueIconSuffix = "/i/leaguelogos/soccer/500/2.png"
	case "eng.1":
		leagueIconSuffix = "/i/leaguelogos/soccer/500/23.png"
	case "eng.fa":
		leagueIconSuffix = "/i/leaguelogos/soccer/500/40.png"
	case "eng.league_cup":
		leagueIconSuffix = "/i/leaguelogos/soccer/500/41.png"
	default:
		leagueIconSuffix = "/guid/119cfa41-71d4-39bf-a790-6273a52b0259/logos/default-dark.png"
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
			League:        league,
			LeagueIconUrl: fmt.Sprintf("https://a.espncdn.com%s", leagueIconSuffix),
			MatchUrl:      event.Link,
		},
		URL: rawURL,
	}, nil
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
