package postgres

import "github.com/jwm1rr0rb10/libraries/backend/golang/queryify"

var (
	SpotTable = queryify.NewTable("public", "spot", "s", "id")
)
