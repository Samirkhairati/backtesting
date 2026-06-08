CREATE TABLE IF NOT EXISTS daily_bars (
    symbol VARCHAR(15) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    open NUMERIC NOT NULL,
    high NUMERIC NOT NULL,
    low NUMERIC NOT NULL,
    close NUMERIC NOT NULL,
    volume BIGINT NOT NULL,
    CONSTRAINT pk_daily_bars PRIMARY KEY (symbol, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_daily_bars_symbol_time ON daily_bars (symbol, timestamp ASC);
