CREATE DATABASE tcc_icmc;

\c tcc_icmc

CREATE TABLE "Signals"
(
    "date"  VARCHAR(20) NOT NULL,
    "agent_id" VARCHAR(10) NOT NULL,
    "temperature"   INT NOT NULL,
    "moisture" INT NOT NULL,
    "state" VARCHAR(3) NOT NULL
    CONSTRAINT "PK_Signals" PRIMARY KEY ("date")
)