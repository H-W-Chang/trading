package server

import (
	"trading/pkg/database"
	"trading/pkg/matcher"
)

type Server interface {
	Serve()
}

func CreateMatcher(matchRule string) matcher.Matcher {
	switch matchRule {
	case "partial":
		return &matcher.PartialMatcher{Ol: &database.MemoryDB{}}
	}
	return nil
}
