package data

import (
	"cli-notes/scripts/config"
	"errors"
)

func GetTeamNames() ([]string, error) {
  names := config.TEAM_NAMES
  
  if len(names) < 1 {
	return nil, errors.New("team names are empty")
  }

  return names, nil
  }