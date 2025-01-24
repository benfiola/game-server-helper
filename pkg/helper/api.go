package helper

import "github.com/benfiola/game-server-helper/pkg/helperapi"

// Wraps [helperapi.Api] - provides helper functions and additional helper related members
type Api struct {
	helperapi.Api
	Directories Map[string, string]
}
