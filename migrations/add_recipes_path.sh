#!/bin/sh

# rename current tables
sqlite3 $1 'alter table `recipes` rename to `recipes_backup`;'

# create tables w/ auto inc
sqlite3 $1 'create table if not exists recipes (id integer primary key autoincrement, url text unique, title text, instructions text, author text, total_time int, yields text, serving_size text, calories text, image blob, path text unique);'

# migrate data
sqlite3 $1 'insert into recipes (url, title, instructions, author, total_time, yields, serving_size, calories, image, path) select url, title, instructions, author, total_time, yields, serving_size, calories, image, id from recipes_backup;'

# drop backups
sqlite3 $1 'drop table recipes_backup;'

# create trigger
sqlite3 $1 '
create trigger if not exists idtopath
after insert on recipes 
begin
  update recipes 
  set path = (
    case
      when new.path is not null then new.path
      else new.id
    end
  )
  where id = new.id;
end'
