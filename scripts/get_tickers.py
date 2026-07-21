import pandas as pd
import json
import requests
import os
from io import StringIO

headers = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'}

# S&P 500
print("Fetching S&P 500 from Wikipedia...")
sp500_url = 'https://en.wikipedia.org/wiki/List_of_S%26P_500_companies'
html = requests.get(sp500_url, headers=headers).text
sp500_table = pd.read_html(StringIO(html))[0]
sp500_tickers = sp500_table['Symbol'].tolist()

# NASDAQ 100
print("Fetching NASDAQ 100 from Slickcharts...")
ndx_url = 'https://www.slickcharts.com/nasdaq100'
html = requests.get(ndx_url, headers=headers).text
ndx_table = pd.read_html(StringIO(html))[0]
ndx_tickers = ndx_table['Symbol'].tolist()

print(f"Saving {len(sp500_tickers)} S&P 500 tickers and {len(ndx_tickers)} NASDAQ 100 tickers...")

# Resolve path dynamically so it works from anywhere
script_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(script_dir)
output_path = os.path.join(project_root, 'client', 'src', 'lib', 'stockLists.ts')

with open(output_path, 'w') as f:
    f.write('export const SP500_TICKERS = ' + json.dumps(sp500_tickers) + ';\n')
    f.write('export const NASDAQ100_TICKERS = ' + json.dumps(ndx_tickers) + ';\n')

print("Done!")
