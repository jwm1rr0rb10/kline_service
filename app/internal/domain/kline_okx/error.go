package kline_okx

import "github.com/jwm1rr0rb10/go-errors"

var (
	ErrKlineAlreadyExist = errors.New("Kline already exist")
)

// -------------------------------------- Errors and constants from storage  --------------------------------------

const (
	SpotKlineIDPKConstraint          = "spot_id_pk"
	SpotSymbolOpenCloseUnqConstraint = "spot_symbol_open_close_unq"
)

var (
	ErrViolatesConstraintSpotIDPK               = errors.New("violates constraint spot_id_pk")
	ErrViolatesConstraintSpotSymbolOpenCloseUnq = errors.New("violates constraint spot_symbol_open_close_unq")
)
