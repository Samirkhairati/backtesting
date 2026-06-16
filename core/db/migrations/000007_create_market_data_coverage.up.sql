CREATE TABLE IF NOT EXISTS market_data_coverage (
    symbol VARCHAR(15) PRIMARY KEY,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
