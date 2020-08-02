/*

 */

package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitspawngg/tournament-bracket-manager/models"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type MatchService struct {
	log *logrus.Entry
	db  *models.DB
}

func getDB() (*models.DB, error) {
	db_type, exists := os.LookupEnv("DB_TYPE")
	if !exists {
		return nil, errors.New("missing DB_TYPE environment variable")
	}
	db_path, exists := os.LookupEnv("DB_PATH")
	if !exists {
		return nil, errors.New("missing DB_PATH environment variable")
	}
	db := models.NewDB(db_type, db_path)
	if err := db.Connect(); err != nil {
		return nil, err
	}
	return db, nil
}

func NewMatchService(log *logrus.Logger, db *models.DB) *MatchService {
	return &MatchService{
		log: log.WithField("services", "Match"),
		db:  db,
	}
}

func GetMatchSchedule(teams []string, format string) ([]models.Match, string, error) {
	// 生成对战表
	// implement proper check for number of teams in the next line
	// 根据传入的队伍名，传出生成的对战表
	lenteam := len(teams)
	if !(lenteam > 0 && lenteam&(lenteam-1) == 0) {
		return nil, "", errors.New("number of teams not a power of 2")
	}
	var matches []models.Match
	uuid4 := uuid.NewV4().String()
	if format == "SINGLE" {
		for i := 0; i < len(teams)/2; i++ {
			matches = append(matches, models.Match{TournamentID: uuid4, Round: 1, Table: i + 1, TeamOne: teams[2*i], TeamTwo: teams[2*i+1], Status: "Pending", Result: 0})
		}
		print(uuid4)
		db, err := getDB()
		if err != nil {
			return nil, "", err
		}
		if err := db.CreateMatches(matches); err != nil {
			return nil, "", err
		}
	} else if format == "CONSOLATION" {
		return nil, "", fmt.Errorf("Unsupported tournament format [%s]", format)
	} else {
		return nil, "", fmt.Errorf("Unsupported tournament format [%s]", format)
	}
	return matches, uuid4, nil
}

func SetMatchResult(tournamentId string, round, table, result int) error {
	// 设置比赛结果
	if result < 1 || result > 2 {
		return errors.New("invalid result")
	}
	db, err := getDB()
	if err != nil {
		return err
	}
	if err := db.SetMatchResult(tournamentId, round, table, result); err != nil {
		// Set Result
		return err
	}
	another_match, err := db.GetPendingMatch(tournamentId, round, table)
	if err != nil && err.Error() == "record not found" {
		return nil
	} else if err != nil {
		return err
	}
	another_winner, err := db.GetMatchWinner(another_match.TournamentID, another_match.Round, another_match.Table)
	if err != nil {
		return err
	}
	winner, err := db.GetMatchWinner(tournamentId, round, table)
	if err != nil {
		return err
	}
	count, err := db.GetRoundCount(tournamentId, round+1)
	if err != nil {
		return err
	}
	if err := db.CreateMatches(
		[]models.Match{
			{
				TournamentID: tournamentId,
				Round:        another_match.Round + 1,
				Table:        count + 1,
				TeamOne:      another_winner,
				TeamTwo:      winner,
				Status:       "Pending",
				Result:       0,
			},
		},
	); err != nil {
		return err
	}
	if db.SetMatchFinished(tournamentId, round, table) == nil && db.SetMatchFinished(another_match.TournamentID, another_match.Round, another_match.Table) == nil {
		return nil
	}
	return errors.New("Set Finished Failed")
}

func GetMatches(tournamentId string) ([]models.Match, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}
	if matches, err := db.GetMatchesByTournament(tournamentId); err != nil {
		return nil, err
	} else {
		return matches, nil
	}
}
