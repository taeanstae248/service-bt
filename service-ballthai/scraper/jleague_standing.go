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
}

	// ScrapeJLeagueStandings scrapes J-League standings from the official J.League site
func ScrapeJLeagueStandings(db *sql.DB) error {
	// Get or create J-League in database
	leagueID, err := getOrCreateLeague(db, "J-League Division 1")
	if err != nil {
		return fmt.Errorf("failed to get or create J-League: %v", err)
	}

	config := JLeagueStandingConfig{
		URL:      "https://www.jleague.co/th/standings/j1/2025/",
		LeagueID: leagueID, // Dynamic J-League ID from database
		Name:     "J-League",
	}

	log.Printf("Scraping J-League standings from: %s", config.URL)

	// Fetch HTML content
	resp, err := http.Get(config.URL)
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

	// Find standings table rows on jleague.co ('.standing-table tbody tr')
	doc.Find(".standing-table tbody tr").Each(func(i int, s *goquery.Selection) {
	       // Extract team data from each row
	       teamData := extractJLeagueTeamData(s)
	       if teamData == nil {
		       return // Skip invalid rows
	       }

	       // Download team logo if needed
	       if teamData.LogoURL != "" {
		       teamData.LogoPath = downloadTeamLogo(teamData.LogoURL)
	       }

	       // Get or create team ID
	       teamID, err := getOrCreateTeamID(db, teamData.Name, teamData.LogoPath, config.LeagueID)
	       if err != nil {
		       log.Printf("Error getting team ID for %s: %v", teamData.Name, err)
		       return
	       }

	       // Prepare standing data
	       standingDB := models.StandingDB{
		       LeagueID:       config.LeagueID,
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
       })

	log.Printf("Completed scraping J-League standings")
	return nil
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

// extractJLeagueTeamData extracts team data from a table row
func extractJLeagueTeamData(s *goquery.Selection) *JLeagueTeamData {
	teamData := &JLeagueTeamData{}

	// Position (column 0)
	positionText := strings.TrimSpace(s.Find("td").Eq(0).Text())
	if pos, err := strconv.Atoi(positionText); err == nil {
		teamData.Position = pos
	}

	// Team name and logo (column 1)
	teamCell := s.Find("td").Eq(1)

	// Prefer the club-item__name text (Thai name) if present
	nameText := strings.TrimSpace(teamCell.Find(".club-item__name").Text())
	if nameText == "" {
		nameText = strings.TrimSpace(teamCell.Text())
	}
	teamData.Name = nameText

	// Extract logo URL from data-src or src attributes on the emblem image
	if logoImg := teamCell.Find(".club-emblem__image").First(); logoImg.Length() > 0 {
		if logoURL, exists := logoImg.Attr("data-src"); exists && logoURL != "" {
			teamData.LogoURL = logoURL
		} else if logoURL, exists := logoImg.Attr("src"); exists {
			teamData.LogoURL = logoURL
		}
	}

	// Handle team name mapping (similar to your PHP)
	teamData.Name = mapJLeagueTeamName(teamData.Name)

	// Extract other statistics
	cells := s.Find("td")
	if cells.Length() >= 10 {
		teamData.Played, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(2).Text()))
		teamData.Wins, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(3).Text()))
		teamData.Draws, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(4).Text()))
		teamData.Losses, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(5).Text()))
		teamData.GoalsFor, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(6).Text()))
		teamData.GoalsAgainst, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(7).Text()))
		teamData.GoalDifference, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(8).Text()))
		teamData.Points, _ = strconv.Atoi(strings.TrimSpace(cells.Eq(9).Text()))
	}

	// Validate that we have essential data
	if teamData.Name == "" || teamData.Position == 0 {
		return nil
	}

	return teamData
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
	// Make root-relative URLs absolute for jleague.co
	if strings.HasPrefix(logoURL, "/") {
		logoURL = "https://www.jleague.co" + logoURL
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
