#!/bin/sh
sqlite3 $1 '
create table if not exists comments (
   id integer primary key, 
   recipeid int references recipes(id), 
   create_timestamp timestamp default current_timestamp, 
   update_timestamp timestamp, 
   who text,
   comment text
);'

sqlite3 $1 '
create trigger comment_update_trigger 
update of comment on comments 
begin
   update comments
   set update_timestamp = current_timestamp
   where id = id;
end;
'
