/*

 */

package models

import (
	"errors"
	_ "github.com/satori/go.uuid"
)

type Match struct {
	TournamentID string `json:"tournamentId" gorm:"primarykey"`
	Round        int    `json:"round" gorm:"primarykey"`
	Table        int    `json:"table" gorm:"primarykey"`
	TeamOne      string `json:"teamOne"`
	TeamTwo      string `json:"teamTwo"`
	Status       string `json:"status"`
	Result       int    `json:"result"` // 1 if Team One wins, 2 if Team Two wins, -1 if no winner
}

func (db DB) CreateMatches(matches []Match) error {
	return db.DB.Create(matches).Error
}

func (db DB) GetMatch(tournamentId string, round, table int) (*Match, error) {
	match := Match{}
	err := db.DB.Where(`"tournament_id" = ? AND "round" = ? AND "table" = ?`, tournamentId, round, table).First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (db DB) GetMatchesByTournament(tournamentId string) ([]Match, error) {
	matches := make([]Match, 0)
	err := db.DB.Order("round").Where("tournament_id = ?", tournamentId).Find(&matches).Error
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (db DB) GetMatchesByStatus(status string) ([]Match, error) {
	matches := make([]Match, 0)
	err := db.DB.Where("status = ?", status).Find(&matches).Error
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (db DB) DeleteMatch(tournamentId string, round, table int) error {
	match := Match{}
	err := db.DB.Where(`"tournament_id" = ? AND "round" = ? AND "table" = ?`, tournamentId, round, table).First(&match).Error
	if err != nil {
		return err
	}
	err = db.DB.Delete(&match).Error
	if err != nil {
		return err
	}
	return nil
}

func (db DB) SetMatchResult(tournamentId string, round, table, result int) error {
	//
	updateMatch := Match{
		Result: result,
	}
	match := Match{}
	if err := db.DB.Where(`"tournament_id" = ? AND "round" = ? AND "table" = ?`, tournamentId, round, table).First(&match).Error; err != nil {
		return err
	}
	if match.Result != 0 {
		return errors.New("Table Finished")
	}
	if err := db.DB.Model(&match).Updates(updateMatch).Error; err != nil {
		return err
	}
	return nil
}

func (db DB) GetMatchWinner(tournamentId string, round, table int) (string, error) {
	match := Match{}
	err := db.DB.Where(`"tournament_id" = ? AND "round" = ? AND "table" = ?`, tournamentId, round, table).First(&match).Error
	if err != nil {
		return "", err
	}
	if match.Result == 0 {
		return "", errors.New("round not finished")
	} else if match.Result == 1 {
		return match.TeamOne, nil
	} else if match.Result == 2 {
		return match.TeamTwo, nil
	}
	return "", errors.New("value in database is not in [0:2]")
}

func (db DB) GetPendingMatch(tournamentId string, round, table int) (Match, error) {
	// 找到另一个已经设置 Result 但是 Status 为 Pending 的
	match := Match{}
	if err := db.DB.Where(
		`"tournament_id" = ? AND "Status" = ? AND "Result" <> ? AND "table" <> ? AND round = ?`,
		tournamentId, "Pending", 0, table, round,
	).First(&match).Error; err != nil {
		return match, err
	}
	return match, nil
}

func (db DB) SetMatchFinished(tournamentId string, round, table int) error {
	// 找到另一个已经设置 Result 但是 Status 为 Pending 的
	match := Match{}
	if err := db.DB.Where(
		`"tournament_id" = ? AND "table" = ? AND round = ?`,
		tournamentId, table, round,
	).First(&match).Error; err != nil {
		return err
	}
	db.DB.Model(&match).Update("Status", "Finished")
	return nil
}

func (db DB) GetRoundCount(tournamentId string, round int) (int, error) {
	var matches []Match
	if err := db.DB.Where(
		`"tournament_id" = ? AND round = ?`,
		tournamentId, round,
	).Find(&matches).Error; err != nil {
		return 0, nil
	}
	return len(matches), nil
}
