# Scrape

Scrape a url for a recipe into sqlite3.

## Requirements

- python 3.0+
- pip

`pip install -r requirements.txt`

## Usage

Will create the database if it doesn't exists.

`python scrape.py [some_sqlite3.db] [some_url]`

Import many from a text file

`while read in; do python3 scrape.py recipes.db "$in"; done < import.txt`
