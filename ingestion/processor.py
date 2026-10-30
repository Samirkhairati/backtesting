import logging
from datetime import date

from downloader import fetch_data
from database import DatabaseHandler

logger = logging.getLogger(__name__)


def missing_segments(start, end, coverage):
    """Date ranges that must still be downloaded for [start, end] given stored coverage.

    Coverage is a single contiguous envelope per symbol, so every segment returned touches
    that envelope. That keeps the envelope truthful when it grows (no hidden gaps).
    """
    if coverage is None:
        return [(start, end)]
    cov_start, cov_end = coverage
    segments = []
    if start < cov_start:
        segments.append((start, cov_start))
    if end > cov_end:
        segments.append((cov_end, end))
    return segments


class JobProcessor:
    def __init__(self, db_url):
        self.db_handler = DatabaseHandler(db_url)

    def process_job(self, job_id, symbol, start_date, end_date):
        """Returns (status, downloaded) where downloaded says whether new bars were stored."""
        try:
            start = date.fromisoformat(str(start_date))
            end = date.fromisoformat(str(end_date))
            downloaded = False
            with self.db_handler.symbol_lock(symbol):
                # Re-read coverage now: another experiment may have fetched this symbol
                # while this job sat in the queue.
                segments = missing_segments(start, end, self.db_handler.get_coverage(symbol))
                if not segments:
                    logger.info(f"{symbol} already covered for {start}..{end}, skipping download")
                    return "COMPLETED", False
                for seg_start, seg_end in segments:
                    if self.db_handler.is_job_cancelled(job_id, "ingestion_jobs"):
                        logger.info(f"Job {job_id} cancelled, stopping early")
                        return "CANCELLED", downloaded
                    rows = fetch_data(symbol, seg_start.isoformat(), seg_end.isoformat())
                    if rows:
                        self.db_handler.insert_data(symbol, seg_start, seg_end, rows)
                        downloaded = True
            return "COMPLETED", downloaded
        except Exception as e:
            logger.error(f"Failed to process {symbol}: {e}")
            return "FAILED", False
