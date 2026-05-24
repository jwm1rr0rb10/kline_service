-- +goose Up
SELECT 'up SQL query';

CREATE TABLE spot (
    id               UUID           NOT NULL, -- UUID for lines 1 min.
    symbol           TEXT           NOT NULL, -- Symbol example: BTCUSDT.
    interval         TEXT           NOT NULL, -- Interval example: 1m, 1H, 1D...
    open_time        TIMESTAMP      NOT NULL, -- Open time for lines 1 min.
    close_time       TIMESTAMP      NOT NULL, -- Close time for lines 1 min.
    type_trend       BOOLEAN        NOT NULL, -- Bearish trend - false, Bullish trend true.
    open_price       NUMERIC(20, 10) NOT NULL, -- Open price for lines 1 min.
    high_price       NUMERIC(20, 10) NOT NULL, -- High price for lines 1 min.
    low_price        NUMERIC(20, 10) NOT NULL, -- Low price for lines 1 min.
    close_price      NUMERIC(20, 10) NOT NULL, -- Close price for lines 1 min.
    base_volume      NUMERIC(20, 10) NOT NULL, -- Volume for lines 1 min.
    quote_volume     NUMERIC(20, 10) NOT NULL, -- Volume quote active.
    CONSTRAINT spot_id_pk PRIMARY KEY (id),
    CONSTRAINT spot_symbol_open_close_unq UNIQUE (symbol, open_time, close_time)
);
-- +goose Down
SELECT 'down SQL query';