import sqlite3
import sys
import requests
from PIL import Image
from io import BytesIO
from recipe_scrapers import scrape_me

# mass import with below
# while read in; do python3 scrape.py recipes.db "$in"; done < import.txt

# get args
db = sys.argv[1]
url = sys.argv[2]

# scrape
scraper = scrape_me(url, wild_mode = True)

# open db
con = sqlite3.connect(db)
cur = con.cursor()

# make tables
cur.execute('''
    create table if not exists recipes
    (id integer primary key, url text unique, title text, instructions text, author text, total_time int, yields text, serving_size text, calories text, image blob)''')

cur.execute('''
    create table if not exists ingredients
    (id integer primary key, recipeid int REFERENCES recipes(id), ingredient text)''')

# generate base64 thumbnail
image = scraper.image()
if image is not None:
    resp = requests.get(image, stream=True)

    if resp.ok:
        im = Image.open(BytesIO(resp.content))
        im.thumbnail((360, 360), Image.Resampling.LANCZOS)

        buffered = BytesIO()
        im.save(buffered, format="webp")
        image = buffered.getvalue()

# dump to db
recipe = (
        None,
        scraper.canonical_url(),
        scraper.title(), 
        scraper.instructions(),
        str(scraper.author()),
        scraper.total_time(),
        scraper.yields(),
        scraper.nutrients().get('servingSize'),
        scraper.nutrients().get('calories'),
        image,
    )

cur.execute("insert into recipes values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", recipe)

ingredients = map(lambda x: (None, cur.lastrowid, x), scraper.ingredients())
cur.executemany("insert into ingredients values(?, ?, ?)", ingredients)

# commit
con.commit()
con.close()
