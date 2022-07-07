#!/bin/sh
sqlite3 $1 "insert into users (email) values('$2');"
