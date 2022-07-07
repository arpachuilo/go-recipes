#!/bin/sh
sqlite3 $1 'create table if not exists users (id integer primary key, email text unique, verification_code text, verification_code_expiry datetime);'
