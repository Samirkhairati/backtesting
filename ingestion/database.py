import logging
from contextlib import contextmanager

import psycopg2
from psycopg2.extras import execute_values

logger = logging.getLogger(__name__)

class DatabaseHandler:
    def __init__(self, db_url):
        self.db_url = db_url

    @contextmanager
    def symbol_lock(self, symbol):
        """Serialise work on one symbol across all ingestion workers.

        Without this, two workers handling the same symbol at the same moment would both
        see "no coverage" and both download. With it, the second waits, then sees the first
        one's coverage and skips the download.
        """
        conn = psycopg2.connect(self.db_url)
        conn.autocommit = True
        cursor = conn.cursor()
        try:
            cursor.execute("SELECT pg_advisory_lock(hashtextextended(%s, 0))", (symbol,))
            yield
        finally:
            try:
                cursor.execute("SELECT pg_advisory_unlock(hashtextextended(%s, 0))", (symbol,))
            finally:
                cursor.close()
                conn.close()

    def get_coverage(self, symbol):
        """Return (start, end) dates already stored for the symbol, or None."""
        conn = psycopg2.connect(self.db_url)
        cursor = conn.cursor()
        try:
            cursor.execute("SELECT start_date, end_date FROM market_data_coverage WHERE symbol = %s", (symbol,))
            row = cursor.fetchone()
            return (row[0], row[1]) if row else None
        finally:
            cursor.close()
            conn.close()

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

    def is_job_cancelled(self, job_id, table="ingestion_jobs"):
        conn = psycopg2.connect(self.db_url)
        cursor = conn.cursor()
        try:
            query = f"SELECT status FROM {table} WHERE id = %s"
            cursor.execute(query, (job_id,))
            res = cursor.fetchone()
            if res and res[0] == 'CANCELLED':
                return True
            return False
        finally:
            cursor.close()
            conn.close()
