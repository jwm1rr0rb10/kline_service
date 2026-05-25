package kline

import (
	"context"

	"github.com/jwm1rr0rb10/go-errors"
	"github.com/jwm1rr0rb10/go-logging"
	gRPCKlineService "github.com/jwm1rr0rb10/kline_contract/gen/go/kline_service/v1"
	policySpotOKX "github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"
)

func (c *Controller) SearchKline(
	ctx context.Context,
	req *gRPCKlineService.SearchKlineRequest,
) (*gRPCKlineService.SearchKlineResponse, error) {
	logging.L(ctx).Debug("SearchKline() called")

	filters, bvfErr := buildValidationSearchKlineFilters(req)
	if bvfErr != nil {
		return nil, errors.Wrap(bvfErr, "buildValidationSearchKlineFilters")
	}

	request := policySpotOKX.NewSearchKlineRequest(
		filters,
		req.WithCount,
	)

	output, err := c.policy.SearchKline(ctx, *request)
	if err != nil {
		return nil, errors.Wrap(err, "c.policy.SearchKline")
	}

	response := convertSearchKlineToPB(output)

	return response, nil
}
