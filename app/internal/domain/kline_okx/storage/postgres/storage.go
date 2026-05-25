package postgres

import (
	"context"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/jwm1rr0rb10/kline_service/app/internal/dal/postgres"
	"github.com/jwm1rr0rb10/kline_service/app/internal/domain"
	"github.com/jwm1rr0rb10/libraries/backend/golang/logging"
	psql "github.com/jwm1rr0rb10/libraries/backend/golang/postgresql"
	"github.com/jwm1rr0rb10/libraries/backend/golang/queryify"
	"github.com/jwm1rr0rb10/libraries/backend/golang/sfqb"
	"github.com/jwm1rr0rb10/libraries/backend/golang/tracing"

	domainKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx"
	domainKlineOKXModel "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/model"
)

type Storage struct {
	qb     squirrel.StatementBuilderType
	client *psql.Client
}

func NewStorage(client *psql.Client) *Storage {
	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &Storage{client: client, qb: qb}
}

func (repo *Storage) All(ctx context.Context, filters sfqb.SFQB) (*[]domainKlineOKXModel.Kline, error) {
	return repo.findBy(ctx, filters)
}

func (repo *Storage) findBy(ctx context.Context, filters sfqb.SFQB) (*[]domainKlineOKXModel.Kline, error) {
	queryify.ApplySearchFilters(filters, domain.TextFormat, domain.Percent)
	queryify.ReplaceFilterLike(filters, domain.ILikeFormat)
	queryify.ReplaceTableToAlias(filters, postgres.SpotTable)

	statement := repo.qb.
		Select(
			"s.open_time",
			"s.close_time",
			"s.open_price",
			"s.close_price",
			"s.high_price",
			"s.low_price",
			"s.base_volume",
			"s.quote_volume",
			"s.id",
			"s.symbol",
			"s.interval",
			"s.type_trend",
		).
		From(postgres.SpotTable.From()).
		Where(filters.Where(), filters.Args()...)

	if limit := filters.Limit(); limit > 0 {
		statement = statement.Limit(uint64(limit))
	}
	if offset := filters.Offset(); offset > 0 {
		statement = statement.Offset(uint64(offset))
	}
	if order := filters.Order(); order != "" {
		statement = statement.OrderBy(order)
	}

	query, args, err := statement.ToSql()
	if err != nil {
		err = psql.ErrCreateQuery(err)
		tracing.Error(ctx, err)
		return nil, err
	}

	tracing.TraceValue(ctx, "sql", query)
	logging.L(ctx).Debug("findBy Kline query", logging.StringAttr("sql", query), logging.AnyAttr("args", args))

	rows, err := repo.client.Query(ctx, query, args...)
	if err != nil {
		err = psql.ErrDoQuery(psql.ParsePgError(err))
		tracing.Error(ctx, err)
		return nil, err
	}
	defer rows.Close()

	kLines := make([]domainKlineOKXModel.Kline, 0, 100)

	for rows.Next() {
		var kline domainKlineOKXModel.Kline

		if err := rows.Scan(
			&kline.StartTime,
			&kline.CloseTime,
			&kline.OpenPrice,
			&kline.ClosePrice,
			&kline.HighPrice,
			&kline.LowPrice,
			&kline.BaseAssetVolume,
			&kline.QuoteAssetVolume,
			&kline.ID,
			&kline.Symbol,
			&kline.Interval,
			&kline.TypeTrend,
		); err != nil {
			err = psql.ErrScan(psql.ParsePgError(err))
			tracing.Error(ctx, err)
			return nil, err
		}

		kLines = append(kLines, kline)
	}

	if err := rows.Err(); err != nil {
		err = psql.ErrScan(psql.ParsePgError(err))
		tracing.Error(ctx, err)
		return nil, err
	}

	return &kLines, nil
}

func (repo *Storage) Create(ctx context.Context, kline *domainKlineOKXModel.Kline) error {
	query, args, err := repo.qb.
		Insert(postgres.SpotTable.String()).
		Columns(
			"id",
			"symbol",
			"interval",
			"open_time",
			"close_time",
			"type_trend",
			"open_price",
			"high_price",
			"low_price",
			"close_price",
			"base_volume",
			"quote_volume",
		).
		Values(
			kline.ID,
			kline.Symbol,
			kline.Interval,
			kline.StartTime,
			kline.CloseTime,
			kline.TypeTrend,
			kline.OpenPrice,
			kline.HighPrice,
			kline.LowPrice,
			kline.ClosePrice,
			kline.BaseAssetVolume,
			kline.QuoteAssetVolume,
		).
		ToSql()
	if err != nil {
		err = psql.ErrCreateQuery(err)
		tracing.Error(ctx, err)

		return err
	}

	tracing.TraceValue(ctx, "sql", query)

	for i, arg := range args {
		tracing.TraceValue(ctx, strconv.Itoa(i), arg)
	}

	_, execErr := repo.client.Exec(ctx, query, args...)
	if execErr != nil {
		if pgErr, ok := psql.IsErrUniqueViolation(execErr); ok {
			switch pgErr.ConstraintName {
			case domainKlineOKX.SpotKlineIDPKConstraint:
				return domainKlineOKX.ErrViolatesConstraintSpotIDPK
			case domainKlineOKX.SpotSymbolOpenCloseUnqConstraint:
				return domainKlineOKX.ErrViolatesConstraintSpotSymbolOpenCloseUnq
			}
		}
		execErr = psql.ErrDoQuery(psql.ParsePgError(execErr))
		return execErr
	}
	return nil
}

func (repo *Storage) Count(ctx context.Context, filters sfqb.SFQB) (uint64, error) {
	queryify.ApplySearchFilters(filters, domain.TextFormat, domain.Percent)
	queryify.ReplaceFilterLike(filters, domain.ILikeFormat)
	queryify.ReplaceTableToAlias(filters, postgres.SpotTable)

	statement := repo.qb.
		Select("count(*)").
		From(postgres.SpotTable.From()).
		Where(filters.Where(), filters.Args()...)

	query, args, err := statement.ToSql()
	if err != nil {
		err = psql.ErrCreateQuery(err)
		tracing.Error(ctx, err)
		return 0, err
	}

	tracing.TraceValue(ctx, "sql", query)
	logging.L(ctx).Debug("Count Kline query", logging.StringAttr("sql", query), logging.AnyAttr("args", args))

	var count uint64
	if err := repo.client.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		err = psql.ErrDoQuery(psql.ParsePgError(err))
		tracing.Error(ctx, err)
		return 0, err
	}

	return count, nil
}
