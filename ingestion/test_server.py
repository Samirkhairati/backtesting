import pytest
from unittest.mock import MagicMock, patch
import pandas as pd
from datetime import datetime
import sys

# Mock grpc generated files before importing server
sys.modules['ingestion_pb2'] = MagicMock()
sys.modules['ingestion_pb2_grpc'] = MagicMock()

from server import run_ingestion

@pytest.fixture
def mock_db():
    with patch('server.psycopg2.connect') as mock_connect:
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_connect.return_value = mock_conn
        mock_conn.cursor.return_value = mock_cursor
        yield mock_conn, mock_cursor

@pytest.fixture
def mock_execute_values():
    with patch('server.execute_values') as mock_ev:
        yield mock_ev

@pytest.fixture
def mock_yf():
    with patch('server.yf.download') as mock_download:
        df = pd.DataFrame({
            'Date': [datetime(2023, 1, 3), datetime(2023, 1, 4)],
            'Open': [130.28, 126.89],
            'High': [130.90, 128.66],
            'Low': [124.17, 125.08],
            'Close': [125.07, 126.36],
            'Volume': [112117500, 89113600]
        })
        df.set_index('Date', inplace=True)
        mock_download.return_value = df
        yield mock_download

def test_run_ingestion_success(mock_db, mock_execute_values, mock_yf):
    mock_conn, mock_cursor = mock_db
    
    run_ingestion("exp-123", ["AAPL"], "2023-01-01", "2023-01-31")
    
    mock_yf.assert_called_once_with("AAPL", start="2023-01-01", end="2023-01-31", progress=False)
    
    # Verify bulk insert
    mock_execute_values.assert_called_once()
    args, kwargs = mock_execute_values.call_args
    assert args[0] == mock_cursor
    assert "INSERT INTO daily_bars" in args[1]
    
    # Check inserted rows mapping
    rows = args[2]
    assert len(rows) == 2
    assert rows[0] == ("AAPL", datetime(2023, 1, 3), 130.28, 130.90, 124.17, 125.07, 112117500)
    assert rows[1] == ("AAPL", datetime(2023, 1, 4), 126.89, 128.66, 125.08, 126.36, 89113600)
    
    # Verify coverage update
    coverage_call = mock_cursor.execute.call_args_list[0]
    assert "INSERT INTO market_data_coverage" in coverage_call[0][0]
    assert coverage_call[0][1] == ("AAPL", "2023-01-01", "2023-01-31")
    
    # Verify experiment status update
    status_call = mock_cursor.execute.call_args_list[1]
    assert "UPDATE experiments SET status = 'INGESTED'" in status_call[0][0]
    assert status_call[0][1] == ("exp-123",)
    
    mock_conn.commit.assert_called_once()
