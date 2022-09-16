package interactor

import "trading/pkg/matcher"

type MatcherFactory interface {
	CreateMatcher(matchRule string) matcher.Matcher
}

func NewMatcher(matchRule string) matcher.Matcher {
	return matcher.CreateMatcher(matchRule)
}
