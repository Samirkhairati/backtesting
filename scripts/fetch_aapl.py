import os
import sys

try:
    import yfinance as yf
except ImportError:
    print("yfinance not installed. Install via: pip install yfinance pandas")
    sys.exit(1)

def main():
    symbol = "AAPL"
    print(f"Fetching 5 years of daily data for {symbol}...")
    
    ticker = yf.Ticker(symbol)
    df = ticker.history(period="5y", interval="1d", auto_adjust=True)
    
    if df.empty:
        print("Error: No data retrieved from yfinance.")
        sys.exit(1)
        
    print(f"Retrieved {len(df)} rows.")

    sql_lines = []
    sql_lines.append("TRUNCATE TABLE daily_bars;")
    sql_lines.append(f"INSERT INTO daily_bars (symbol, timestamp, open, high, low, close, volume) VALUES")
    
    values = []
    for index, row in df.iterrows():
        ts = index.strftime('%Y-%m-%d 00:00:00Z')
        op = round(row['Open'], 4)
        hi = round(row['High'], 4)
        lo = round(row['Low'], 4)
        cl = round(row['Close'], 4)
        vo = int(row['Volume'])
        
        values.append(f"('{symbol}', '{ts}', {op}, {hi}, {lo}, {cl}, {vo})")

    sql_lines.append(",\n".join(values))
    sql_lines.append("ON CONFLICT (symbol, timestamp) DO UPDATE")
    sql_lines.append("SET open = EXCLUDED.open, high = EXCLUDED.high, low = EXCLUDED.low, close = EXCLUDED.close, volume = EXCLUDED.volume;")

    
    script_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(script_dir)
    migrations_dir = os.path.join(project_root, "engine", "db", "migrations")
    
    if not os.path.exists(migrations_dir):
        os.makedirs(migrations_dir)
        
    out_file = os.path.join(migrations_dir, "000003_seed_aapl_5y.up.sql")
    
    with open(out_file, "w") as f:
        f.write("\n".join(sql_lines) + "\n")
        
    print(f"Successfully wrote migration file to: {out_file}")

if __name__ == "__main__":
    main()
