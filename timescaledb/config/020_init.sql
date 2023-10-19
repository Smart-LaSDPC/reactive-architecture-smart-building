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

SELECT create_hypertable('tbl_temperature_moisture', 'time');

