package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-ballthai-scraper/database"
	"go-ballthai-scraper/models"

	"github.com/PuerkitoBio/goquery"
)

// JLeagueStandingConfig configuration for J-League standings scraping
type JLeagueStandingConfig struct {
	URL      string
	LeagueID int
	Name     string
	StageName string
}

// ScrapeJLeagueStandings scrapes J-League standings from both EAST and WEST stages
func ScrapeJLeagueStandings(db *sql.DB) error {
	// Get or create J-League in database
	leagueID, err := getOrCreateLeague(db, "J-League Division 1")
	if err != nil {
		return fmt.Errorf("failed to get or create J-League: %v", err)
	}

	// Scrape both EAST and WEST stages
	stages := []struct {
		name string
		url  string
	}{
		{"east", "https://www.thscore.mobi/football/database/league-25/3540"},
		{"west", "https://www.thscore.mobi/football/database/league-25/3541"},
	}

	for _, stage := range stages {
		if err := scrapeJLeagueStandingsByStage(db, leagueID, stage.name, stage.url); err != nil {
			log.Printf("Error scraping %s stage: %v", stage.name, err)
		}
	}

	log.Printf("Completed scraping J-League standings for all stages")
	return nil
}

// scrapeJLeagueStandingsByStage scrapes J-League standings for a specific stage
func scrapeJLeagueStandingsByStage(db *sql.DB, leagueID int, stageName, url string) error {

	log.Printf("Scraping J-League standings for %s stage from: %s", stageName, url)

	// Get or create stage
	stageID, err := getOrCreateStage(db, stageName)
	if err != nil {
		return fmt.Errorf("failed to get or create stage %s: %v", stageName, err)
	}

	// Fetch HTML content
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch J-League standings: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to fetch J-League standings: status code %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Extract team names and logos from the first table in .rankbox
	var teamsList []map[string]string
	doc.Find(".rankbox table.eTable tbody tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if s.Find("th").Length() > 0 {
			return
		}

		cells := s.Find("td")
		if cells.Length() < 3 {
			return
		}

		// Extract rank
		rank := strings.TrimSpace(cells.Eq(0).Find("span.whiteTxt").Text())

		// Extract team logo URL
		logoURL := ""
		if img := cells.Eq(1).Find("img.teamIcon"); img.Length() > 0 {
			if src, exists := img.Attr("src"); exists {
				logoURL = src
			}
		}

		// Extract team name
		teamName := strings.TrimSpace(cells.Eq(2).Find("a.LName").Text())

		if teamName != "" && rank != "" {
			teamsList = append(teamsList, map[string]string{
				"rank":     rank,
				"name":     teamName,
				"logoURL":  logoURL,
			})
		}
	})

	// Extract statistics from the second table in .rankdata
	var statsList []map[string]string
	doc.Find(".rankdata table.eTable tbody tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if s.Find("th").Length() > 0 {
			return
		}

		cells := s.Find("td")
		if cells.Length() < 13 {
			return
		}

		// Extract statistics
		played := strings.TrimSpace(cells.Eq(0).Text())
		wins := strings.TrimSpace(cells.Eq(1).Text())
		draws := strings.TrimSpace(cells.Eq(2).Text())
		losses := strings.TrimSpace(cells.Eq(3).Text())
		points := strings.TrimSpace(cells.Eq(4).Text())
		goalsFor := strings.TrimSpace(cells.Eq(5).Text())
		goalsAgainst := strings.TrimSpace(cells.Eq(6).Text())
		goalDiff := strings.TrimSpace(cells.Eq(7).Text())

		statsList = append(statsList, map[string]string{
			"played":       played,
			"wins":         wins,
			"draws":        draws,
			"losses":       losses,
			"points":       points,
			"goalsFor":     goalsFor,
			"goalsAgainst": goalsAgainst,
			"goalDiff":     goalDiff,
		})
	})

	// Match teams with their statistics by position
	for pos := 0; pos < len(teamsList) && pos < len(statsList); pos++ {
		teamInfo := teamsList[pos]
		statsInfo := statsList[pos]

		teamData := &JLeagueTeamData{
			Position:       pos + 1,
			Name:           teamInfo["name"],
			LogoURL:        teamInfo["logoURL"],
			Played:         atoi(statsInfo["played"]),
			Wins:           atoi(statsInfo["wins"]),
			Draws:          atoi(statsInfo["draws"]),
			Losses:         atoi(statsInfo["losses"]),
			Points:         atoi(statsInfo["points"]),
			GoalsFor:       atoi(statsInfo["goalsFor"]),
			GoalsAgainst:   atoi(statsInfo["goalsAgainst"]),
			GoalDifference: atoi(statsInfo["goalDiff"]),
		}

		// Download team logo if needed
		if teamData.LogoURL != "" {
			teamData.LogoPath = downloadTeamLogo(teamData.LogoURL)
		}

		// Get or create team ID
		teamID, err := getOrCreateTeamID(db, teamData.Name, teamData.LogoPath, leagueID)
		if err != nil {
			log.Printf("Error getting team ID for %s: %v", teamData.Name, err)
			continue
		}

		// Prepare standing data
		standingDB := models.StandingDB{
			LeagueID:       leagueID,
			TeamID:         teamID,
			MatchesPlayed:  teamData.Played,
			Wins:           teamData.Wins,
			Draws:          teamData.Draws,
			Losses:         teamData.Losses,
			GoalsFor:       teamData.GoalsFor,
			GoalsAgainst:   teamData.GoalsAgainst,
			GoalDifference: teamData.GoalDifference,
			Points:         teamData.Points,
			CurrentRank:    sql.NullInt64{Int64: int64(teamData.Position), Valid: teamData.Position != 0},
			StageID:        sql.NullInt64{Int64: stageID, Valid: true}, // เพิ่ม stage_id
			Status:         sql.NullInt64{Valid: false}, // เพิ่ม status (null)
		}

		// Insert or update standing in database
		err = database.InsertOrUpdateStanding(db, standingDB)
		if err != nil {
			log.Printf("Error saving standing for team %s: %v", teamData.Name, err)
		} else {
			log.Printf("Successfully saved standing for %s (Position: %d, Points: %d)",
				teamData.Name, teamData.Position, teamData.Points)
		}
	}

	return nil
}

// getOrCreateStage gets stage ID or creates new stage if not exists
func getOrCreateStage(db *sql.DB, stageName string) (int64, error) {
	// Try to find existing stage
	var stageID int64
	query := `SELECT id FROM stage WHERE stage_name = ? LIMIT 1`
	err := db.QueryRow(query, stageName).Scan(&stageID)

	if err == nil {
		log.Printf("Found existing stage: %s (ID: %d)", stageName, stageID)
		return stageID, nil // Stage found
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("error searching for stage: %v", err)
	}

	// Stage not found, create new one
	insertQuery := `INSERT INTO stage (stage_name) VALUES (?)`
	result, err := db.Exec(insertQuery, stageName)
	if err != nil {
		return 0, fmt.Errorf("error creating new stage: %v", err)
	}

	newStageID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting new stage ID: %v", err)
	}

	log.Printf("Created new stage: %s (ID: %d)", stageName, newStageID)
	return newStageID, nil
}

// JLeagueTeamData represents team data extracted from the table
type JLeagueTeamData struct {
	Position       int
	Name           string
	LogoURL        string
	LogoPath       string
	Played         int
	Wins           int
	Draws          int
	Losses         int
	GoalsFor       int
	GoalsAgainst   int
	GoalDifference int
	Points         int
}

// atoi converts string to int, returns 0 if conversion fails
func atoi(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}


// mapJLeagueTeamName maps team names (similar to your PHP logic)
func mapJLeagueTeamName(name string) string {
	// Trim spaces and apply mappings
	name = strings.TrimSpace(name)

	// Team name mappings (same as your PHP)
	nameMap := map[string]string{
		"AVISPA FUKUOKA":   "อวิสป้า ฟูกุโอกะ",
		"TOKUSHIMA VORTIS": "โทคุชิมะ วอร์ติส",
		"KYOTO SANGA":      "เกียวโต แซงก้า",
	}

	if mappedName, exists := nameMap[name]; exists {
		return mappedName
	}

	return name
}

// downloadTeamLogo downloads team logo image (similar to your PHP getImages function)
func downloadTeamLogo(logoURL string) string {
	if logoURL == "" {
		return ""
	}

	// Fix relative URLs by adding https scheme
	if strings.HasPrefix(logoURL, "//") {
		logoURL = "https:" + logoURL
	}
	// Make root-relative URLs absolute for different sources
	if strings.HasPrefix(logoURL, "/") {
		if strings.Contains(logoURL, "football.thscore") {
			logoURL = "https://football.thscore2.com" + logoURL
		} else {
			logoURL = "https://www.jleague.co" + logoURL
		}
	}

	// Create logo path into img/teams
	filename := filepath.Base(logoURL)
	logoPath := "/img/teams/" + filename
	fullPath := filepath.Join(".", "img", "teams", filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return logoPath
	}

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return logoPath // File already exists
	}

	// Download the image
	resp, err := http.Get(logoURL)
	if err != nil {
		log.Printf("Error downloading logo from %s: %v", logoURL, err)
		return logoPath
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Error downloading logo: status code %d", resp.StatusCode)
		return logoPath
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Error creating logo file %s: %v", fullPath, err)
		return logoPath
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Printf("Error saving logo file %s: %v", fullPath, err)
		return logoPath
	}

	log.Printf("Successfully downloaded logo: %s", fullPath)
	return logoPath
}

// getOrCreateLeague gets league ID or creates new league if not exists
func getOrCreateLeague(db *sql.DB, leagueName string) (int, error) {
	// Try to find existing league
	var leagueID int
	query := `SELECT id FROM leagues WHERE name = ? LIMIT 1`
	err := db.QueryRow(query, leagueName).Scan(&leagueID)

	if err == nil {
		log.Printf("Found existing league: %s (ID: %d)", leagueName, leagueID)
		return leagueID, nil // League found
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("error searching for league: %v", err)
	}

	// League not found, create new one
	insertQuery := `INSERT INTO leagues (name) VALUES (?)`
	result, err := db.Exec(insertQuery, leagueName)
	if err != nil {
		return 0, fmt.Errorf("error creating new league: %v", err)
	}

	newLeagueID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting new league ID: %v", err)
	}

	log.Printf("Created new league: %s (ID: %d)", leagueName, newLeagueID)
	return int(newLeagueID), nil
}

// getOrCreateTeamID gets team ID or creates new team (similar to your PHP getTeamId function)
func getOrCreateTeamID(db *sql.DB, teamName, logoPath string, leagueID int) (int, error) {
	// Try to find existing team
	var teamID int
	query := `SELECT id FROM teams WHERE REPLACE(name_th, ' ', '') = REPLACE(?, ' ', '') ORDER BY id DESC LIMIT 1`
	err := db.QueryRow(query, teamName).Scan(&teamID)

	if err == nil {
		return teamID, nil // Team found
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("error searching for team: %v", err)
	}

	// Team not found, create new one
	insertQuery := `INSERT INTO teams (name_th, logo_url) VALUES (?, ?)`
	result, err := db.Exec(insertQuery, teamName, logoPath)
	if err != nil {
		return 0, fmt.Errorf("error creating new team: %v", err)
	}

	newTeamID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting new team ID: %v", err)
	}

	log.Printf("Created new team: %s (ID: %d)", teamName, newTeamID)
	return int(newTeamID), nil
}
