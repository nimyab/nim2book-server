-- +goose Up
-- +goose StatementBegin
create table if not exists dictionary_examples
(
    id                  uuid primary key default uuid_generate_v4(),
    text                text not null,
    translated_text     text not null,
    word_position_start int  not null,
    word_position_end   int  not null,
    dictionary_id       uuid not null,
    constraint fk_dictionary_id foreign key (dictionary_id) references dictionary (id) on delete cascade
);

alter table dictionary
    drop column if exists lang,
    drop column if exists content,
    add column from_lang_code varchar(10)    not null,
    add column to_lang_code   varchar(10)    not null,
    add column part_of_speech varchar(50)    not null,
    add column translations   varchar(255)[] not null,
    add column transcription  varchar(255),
    alter column text type varchar(255),
    add constraint text_part_of_speech_unique_key unique (text, part_of_speech);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists dictionary_examples;

alter table dictionary
    drop constraint if exists text_part_of_speech_unique_key,
    alter column text type text,
    drop column if exists transcription,
    drop column if exists translations,
    drop column if exists part_of_speech,
    drop column if exists to_lang_code,
    drop column if exists from_lang_code,
    add column lang    varchar(10),
    add column content jsonb;
-- +goose StatementEnd
