import time
import logging
import yfinance as yf
import pandas as pd

logger = logging.getLogger(__name__)

def fetch_data(symbol, start_date, end_date):
    logger.info(f"Downloading {symbol} from {start_date} to {end_date}")
    time.sleep(1) # Rate limit protection
    
    data = yf.download(symbol, start=start_date, end=end_date, progress=False)
    
    if data.empty:
        logger.warning(f"No data found for {symbol}")
        return []
        
    data = data.reset_index()
    
    rows = []
    for _, row in data.iterrows():
        def get_val(col_name):
            val = row[col_name]
            return val.iloc[0] if isinstance(val, pd.Series) else val

        date_val = get_val('Date')
        open_val = get_val('Open')
        
        if pd.isna(open_val):
            continue
            
        high_val = get_val('High')
        low_val = get_val('Low')
        close_val = get_val('Close')
        vol_val = get_val('Volume')
        
        dt = date_val.to_pydatetime()
        rows.append((symbol, dt, float(open_val), float(high_val), float(low_val), float(close_val), int(vol_val)))
        
    return rows
