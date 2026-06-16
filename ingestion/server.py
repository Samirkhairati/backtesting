import sys
import os
import time
import threading
import logging
from concurrent import futures

import grpc
import psycopg2
from psycopg2.extras import execute_values
import yfinance as yf
import pandas as pd
from datetime import datetime

import ingestion_pb2
import ingestion_pb2_grpc

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

DB_URL = os.environ.get("DATABASE_URL")
if not DB_URL:
    raise ValueError("DATABASE_URL environment variable is not set and is required")

def run_ingestion(experiment_id, symbols, start_date, end_date):
    logger.info(f"Starting ingestion for {experiment_id} with symbols {symbols}")
    
    try:
        conn = psycopg2.connect(DB_URL)
        cursor = conn.cursor()
        
        for symbol in symbols:
            logger.info(f"Downloading {symbol} from {start_date} to {end_date}")
            time.sleep(1) # Rate limit protection
            
            # Download data
            data = yf.download(symbol, start=start_date, end=end_date, progress=False)
            
            if data.empty:
                logger.warning(f"No data found for {symbol}")
                continue
                
            # Reset index to access 'Date'
            data = data.reset_index()
            
            # Prepare rows
            rows = []
            for _, row in data.iterrows():
                # Helper to extract scalar value from potentially MultiIndex Series
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
                
            if rows:
                # Bulk insert
                insert_query = """
                    INSERT INTO daily_bars (symbol, timestamp, open, high, low, close, volume)
                    VALUES %s
                    ON CONFLICT (symbol, timestamp) DO NOTHING
                """
                execute_values(cursor, insert_query, rows)
                
                # Update coverage based on requested dates, not just what was found
                # This prevents it from continually searching for data that doesn't exist
                coverage_query = """
                    INSERT INTO market_data_coverage (symbol, start_date, end_date)
                    VALUES (%s, %s, %s)
                    ON CONFLICT (symbol) DO UPDATE 
                    SET start_date = LEAST(market_data_coverage.start_date, EXCLUDED.start_date),
                        end_date = GREATEST(market_data_coverage.end_date, EXCLUDED.end_date),
                        last_updated = CURRENT_TIMESTAMP
                """
                cursor.execute(coverage_query, (symbol, start_date, end_date))
                
        # Update experiment
        cursor.execute("UPDATE experiments SET status = 'INGESTED' WHERE id = %s", (experiment_id,))
        conn.commit()
        logger.info(f"Ingestion completed successfully for {experiment_id}")
        
    except Exception as e:
        logger.error(f"Ingestion failed: {e}")
        if 'conn' in locals():
            conn.rollback()
    finally:
        if 'cursor' in locals():
            cursor.close()
        if 'conn' in locals():
            conn.close()

class IngestionService(ingestion_pb2_grpc.IngestionServiceServicer):
    def IngestData(self, request, context):
        logger.info(f"Received IngestData request for {request.experiment_id}")
        
        # Run in background to not block coordinator
        thread = threading.Thread(
            target=run_ingestion, 
            args=(request.experiment_id, list(request.symbols), request.start_date, request.end_date)
        )
        thread.start()
        
        return ingestion_pb2.IngestDataResponse(success=True)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ingestion_pb2_grpc.add_IngestionServiceServicer_to_server(IngestionService(), server)
    port = "50052"
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"Ingestion service started on port {port}")
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
