package server

import (
	"log"
	"trading/pkg/matcher"
)

type Server interface {
	Serve()
}

var or matcher.OrderRepository

func NewServer(serverType string, repo matcher.OrderRepository) Server {
	or = repo
	switch serverType {
	case "http":
		return &HttpServer{}
	default:
		log.Println("Server type not supported")
		return nil
	}
}

func GetMatcher(matchRule string) matcher.Matcher {
	switch matchRule {
	case "partial":
		return &matcher.PartialMatcher{Or: or}
	}
	return nil
}
