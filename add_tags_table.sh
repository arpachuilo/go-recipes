#!/bin/sh
sqlite3 $1 'create table if not exists tags (id integer primary key, recipeid int REFERENCES recipes(id), tag text);'
