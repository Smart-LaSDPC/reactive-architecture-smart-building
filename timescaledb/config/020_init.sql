CREATE DATABASE tcc_icmc;

\c tcc_icmc

CREATE TABLE tbl_temperature_moisture
(
    time        TIMESTAMPTZ     NOT NULL,
    agent_id    TEXT            NOT NULL,
    state       TEXT            NULL,
    temperature INTEGER         NULL,
    moisture    INTEGER         NULL
);

SELECT create_hypertable('tbl_temperature_moisture', 'time', chunk_time_interval => INTERVAL '7 days');

-- CHUNK_TIME_INTEVARL: A documentação recomenda definir os chunks como sendo de no máximo 25% da capa-
--                      cidade de RAM disponível para determinado o periodo de tempo que ele ocupe.
--                      Ex.: 1 GB de dados são produzidos por dia, supondo que tenhamos 16 GB de RAM 
--                      disponível, a duração do chunk aconselhável seria de 4 dias.
--                                      4 GB produzido em 4 dias / 16 GB total = 25%
--                      DEFAULT: Intervalo de 7 dias