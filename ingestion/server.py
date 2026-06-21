import sys
import os
import time
import json
import logging
from datetime import datetime

import yfinance as yf
import psycopg2
from psycopg2.extras import execute_values
import pandas as pd
from kafka import KafkaConsumer, KafkaProducer

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

DB_URL = os.environ.get("DATABASE_URL")
if not DB_URL:
    raise ValueError("DATABASE_URL environment variable is not set and is required")

KAFKA_BROKERS = os.environ.get("KAFKA_BROKERS", "kafka:9092").split(",")

def process_job(symbol, start_date, end_date):
    logger.info(f"Downloading {symbol} from {start_date} to {end_date}")
    time.sleep(1) # Rate limit protection
    
    conn = psycopg2.connect(DB_URL)
    cursor = conn.cursor()
    try:
        data = yf.download(symbol, start=start_date, end=end_date, progress=False)
        
        if data.empty:
            logger.warning(f"No data found for {symbol}")
            return "COMPLETED"
            
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
            
        if rows:
            insert_query = """
                INSERT INTO daily_bars (symbol, timestamp, open, high, low, close, volume)
                VALUES %s
                ON CONFLICT (symbol, timestamp) DO NOTHING
            """
            execute_values(cursor, insert_query, rows)
            
            coverage_query = """
                INSERT INTO market_data_coverage (symbol, start_date, end_date)
                VALUES (%s, %s, %s)
                ON CONFLICT (symbol) DO UPDATE 
                SET start_date = LEAST(market_data_coverage.start_date, EXCLUDED.start_date),
                    end_date = GREATEST(market_data_coverage.end_date, EXCLUDED.end_date),
                    last_updated = CURRENT_TIMESTAMP
            """
            cursor.execute(coverage_query, (symbol, start_date, end_date))
            
        conn.commit()
        return "COMPLETED"
    except Exception as e:
        logger.error(f"Failed to process {symbol}: {e}")
        conn.rollback()
        return "FAILED"
    finally:
        cursor.close()
        conn.close()

if __name__ == '__main__':
    logger.info("Starting ingestion Kafka consumer...")
    
    # Wait for Kafka to be ready
    producer = None
    for _ in range(10):
        try:
            producer = KafkaProducer(
                bootstrap_servers=KAFKA_BROKERS,
                value_serializer=lambda v: json.dumps(v).encode('utf-8')
            )
            break
        except Exception as e:
            logger.warning(f"Waiting for Kafka: {e}")
            time.sleep(3)
            
    if not producer:
        logger.error("Could not connect to Kafka producer")
        sys.exit(1)

    # Allow time for topics to be created by kafka-init container
    time.sleep(5)

    consumer = KafkaConsumer(
        'ingestion-jobs',
        bootstrap_servers=KAFKA_BROKERS,
        group_id='ingestion-worker-group',
        value_deserializer=lambda m: json.loads(m.decode('utf-8')),
        enable_auto_commit=False,
        auto_offset_reset='earliest'
    )
    
    for message in consumer:
        job = message.value
        job_id = job.get('job_id')
        exp_id = job.get('experiment_id')
        symbol = job.get('symbol')
        start = job.get('start_date')
        end = job.get('end_date')
        
        logger.info(f"Processing job {job_id} for {symbol}")
        
        status = process_job(symbol, start, end)
        
        producer.send('ingestion-completed-jobs', {
            'job_id': job_id,
            'experiment_id': exp_id,
            'symbol': symbol,
            'status': status
        })
        producer.flush()
        
        # Commit only after successful processing and publishing
        consumer.commit()
