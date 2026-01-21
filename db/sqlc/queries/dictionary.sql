-- name: GetDictionaryWord :one
select id, text, from_lang_code, to_lang_code, part_of_speech, translations, transcription
from dictionary
where text = sqlc.arg(text)
  and from_lang_code = sqlc.arg(from_lang_code)
  and to_lang_code = sqlc.arg(to_lang_code)
  and part_of_speech = sqlc.arg(part_of_speech);

-- name: GetDictionaryWordById :one
select id, text, from_lang_code, to_lang_code, part_of_speech, translations, transcription
from dictionary
where id = sqlc.arg(id);

-- name: DictionaryWordExists :one
select exists(select 1
              from dictionary
              where text = sqlc.arg(text)
                and from_lang_code = sqlc.arg(from_lang_code)
                and to_lang_code = sqlc.arg(to_lang_code)
                and part_of_speech = sqlc.arg(part_of_speech));

-- name: CreateDictionaryWord :one
insert into dictionary (text, from_lang_code, to_lang_code, part_of_speech, translations, transcription)
values (sqlc.arg(text),
        sqlc.arg(from_lang_code),
        sqlc.arg(to_lang_code),
        sqlc.arg(part_of_speech),
        sqlc.arg(translations),
        sqlc.narg(transcription))
returning id;

-- name: UpdateDictionaryWord :exec
update dictionary
set translations  = sqlc.arg(translations),
    transcription = sqlc.narg(transcription)
where id = sqlc.arg(id);

-- name: DeleteDictionaryWord :exec
delete
from dictionary
where id = sqlc.arg(id);

-- name: SearchDictionaryWords :many
select id, text, from_lang_code, to_lang_code, part_of_speech, translations, transcription
from dictionary
where text ilike sqlc.arg(search_text)
  and from_lang_code = sqlc.arg(from_lang_code)
  and to_lang_code = sqlc.arg(to_lang_code)
order by text
limit sqlc.arg(limit_count) offset sqlc.arg(offset_count);

-- name: GetDictionaryWordsByText :many
select id, text, from_lang_code, to_lang_code, part_of_speech, translations, transcription
from dictionary
where text = sqlc.arg(text)
  and from_lang_code = sqlc.arg(from_lang_code)
  and to_lang_code = sqlc.arg(to_lang_code)
order by part_of_speech;

-- Dictionary Examples queries

-- name: CreateDictionaryExample :one
insert into dictionary_examples (text, translated_text, word_position_start, word_position_end, dictionary_id)
values (sqlc.arg(text),
        sqlc.arg(translated_text),
        sqlc.arg(word_position_start),
        sqlc.arg(word_position_end),
        sqlc.arg(dictionary_id))
returning id;

-- name: GetDictionaryExamples :many
select id, text, translated_text, word_position_start, word_position_end, dictionary_id
from dictionary_examples
where dictionary_id = sqlc.arg(dictionary_id)
order by id;

-- name: GetDictionaryExampleById :one
select id, text, translated_text, word_position_start, word_position_end, dictionary_id
from dictionary_examples
where id = sqlc.arg(id);

-- name: UpdateDictionaryExample :exec
update dictionary_examples
set text                = sqlc.arg(text),
    translated_text     = sqlc.arg(translated_text),
    word_position_start = sqlc.arg(word_position_start),
    word_position_end   = sqlc.arg(word_position_end)
where id = sqlc.arg(id);

-- name: DeleteDictionaryExample :exec
delete
from dictionary_examples
where id = sqlc.arg(id);

-- name: DeleteDictionaryExamplesByWordId :exec
delete
from dictionary_examples
where dictionary_id = sqlc.arg(dictionary_id);

-- name: GetDictionaryWordWithExamples :many
select d.id                   as dictionary_id,
       d.text                 as dictionary_text,
       d.from_lang_code       as dictionary_from_lang_code,
       d.to_lang_code         as dictionary_to_lang_code,
       d.part_of_speech       as dictionary_part_of_speech,
       d.translations         as dictionary_translations,
       d.transcription        as dictionary_transcription,
       de.id                  as example_id,
       de.text                as example_text,
       de.translated_text     as example_translated_text,
       de.word_position_start as example_word_position_start,
       de.word_position_end   as example_word_position_end
from dictionary d
         left join dictionary_examples de on d.id = de.dictionary_id
where d.id = sqlc.arg(dictionary_id)
order by de.id;
