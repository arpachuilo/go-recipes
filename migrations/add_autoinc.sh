#!/bin/sh
# rename current tables
sqlite3 $1 'alter table `recipes` rename to `recipes_backup`;'
sqlite3 $1 'alter table `ingredients` rename to `ingredients_backup`;'
sqlite3 $1 'alter table `tags` rename to `tags_backup`;'

# create tables w/ auto inc
sqlite3 $1 'create table if not exists recipes (id integer primary key autoincrement, url text unique, title text, instructions text, author text, total_time int, yields text, serving_size text, calories text, image blob);'

sqlite3 $1 'create table if not exists ingredients (id integer primary key autoincrement, recipeid int REFERENCES recipes(id), ingredient text);'

sqlite3 $1 'create table if not exists tags (id integer primary key autoincrement, recipeid int REFERENCES recipes(id), tag text);'

# migrate data
sqlite3 $1 'insert into recipes (url, title, instructions, author, total_time, yields, serving_size, calories, image) select url, title, instructions, author, total_time, yields, serving_size, calories, image from recipes_backup;'

sqlite3 $1 'insert into ingredients (recipeid, ingredient) select recipeid, ingredient from ingredients_backup;'

sqlite3 $1 'insert into tags (recipeid, tag) select recipeid, tag from tags_backup;'

# drop backups
sqlite3 $1 'drop table recipes_backup;'
sqlite3 $1 'drop table ingredients_backup;'
sqlite3 $1 'drop table tags_backup;'
