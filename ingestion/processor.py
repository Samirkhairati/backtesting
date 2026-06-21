import logging
from downloader import fetch_data
from database import DatabaseHandler

logger = logging.getLogger(__name__)

class JobProcessor:
    def __init__(self, db_url):
        self.db_handler = DatabaseHandler(db_url)

    def process_job(self, symbol, start_date, end_date):
        try:
            rows = fetch_data(symbol, start_date, end_date)
            if rows:
                self.db_handler.insert_data(symbol, start_date, end_date, rows)
            return "COMPLETED"
        except Exception as e:
            logger.error(f"Failed to process {symbol}: {e}")
            return "FAILED"
