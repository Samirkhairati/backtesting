import psycopg2
from psycopg2.extras import execute_values
import logging

logger = logging.getLogger(__name__)

class DatabaseHandler:
    def __init__(self, db_url):
        self.db_url = db_url

    def insert_data(self, symbol, start_date, end_date, rows):
        if not rows:
            return

        conn = psycopg2.connect(self.db_url)
        cursor = conn.cursor()
        try:
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
        except Exception as e:
            conn.rollback()
            raise e
        finally:
            cursor.close()
            conn.close()
