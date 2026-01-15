-- name: GetDictionaryData :one
select content
from dictionary
where text = $1
  and lang = $2;

-- name: DictionaryDataExists :one
select exists(
    select id
    from dictionary
    where text = $1
      and lang = $2
  );

-- name: CreateDictionaryData :exec
insert into dictionary (text, lang, content)
values ($1, $2, $3);
