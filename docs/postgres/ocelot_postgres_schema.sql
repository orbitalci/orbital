--
-- PostgreSQL database dump
--

-- Dumped from database version 10.1
-- Dumped by pg_dump version 10.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: postgres; Type: COMMENT; Schema: -; Owner: postgres
--

COMMENT ON DATABASE postgres IS 'default administrative connection database';


--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: build_failure_reason; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE build_failure_reason (
    build_id integer NOT NULL,
    reasons jsonb,
    id integer NOT NULL
);


ALTER TABLE build_failure_reason OWNER TO postgres;

--
-- Name: build_failure_reason_build_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE build_failure_reason_build_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE build_failure_reason_build_id_seq OWNER TO postgres;

--
-- Name: build_failure_reason_build_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE build_failure_reason_build_id_seq OWNED BY build_failure_reason.build_id;


--
-- Name: build_failure_reason_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE build_failure_reason_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE build_failure_reason_id_seq OWNER TO postgres;

--
-- Name: build_failure_reason_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE build_failure_reason_id_seq OWNED BY build_failure_reason.id;


--
-- Name: build_output; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE build_output (
    build_id integer NOT NULL,
    output character varying,
    id integer NOT NULL
);


ALTER TABLE build_output OWNER TO postgres;

--
-- Name: build_output_build_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE build_output_build_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE build_output_build_id_seq OWNER TO postgres;

--
-- Name: build_output_build_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE build_output_build_id_seq OWNED BY build_output.build_id;


--
-- Name: build_output_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE build_output_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE build_output_id_seq OWNER TO postgres;

--
-- Name: build_output_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE build_output_id_seq OWNED BY build_output.id;


--
-- Name: build_summary; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE build_summary (
    hash character varying(50),
    failed boolean,
    starttime timestamp without time zone,
    account character varying(50),
    buildtime numeric,
    repo character varying(100),
    id integer NOT NULL
);


ALTER TABLE build_summary OWNER TO postgres;

--
-- Name: build_summary_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE build_summary_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE build_summary_id_seq OWNER TO postgres;

--
-- Name: build_summary_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE build_summary_id_seq OWNED BY build_summary.id;


--
-- Name: build_failure_reason build_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_failure_reason ALTER COLUMN build_id SET DEFAULT nextval('build_failure_reason_build_id_seq'::regclass);


--
-- Name: build_failure_reason id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_failure_reason ALTER COLUMN id SET DEFAULT nextval('build_failure_reason_id_seq'::regclass);


--
-- Name: build_output build_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_output ALTER COLUMN build_id SET DEFAULT nextval('build_output_build_id_seq'::regclass);


--
-- Name: build_output id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_output ALTER COLUMN id SET DEFAULT nextval('build_output_id_seq'::regclass);


--
-- Name: build_summary id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_summary ALTER COLUMN id SET DEFAULT nextval('build_summary_id_seq'::regclass);


--
-- Name: build_failure_reason build_failure_reason_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_failure_reason
    ADD CONSTRAINT build_failure_reason_pkey PRIMARY KEY (id);


--
-- Name: build_output build_output_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_output
    ADD CONSTRAINT build_output_pkey PRIMARY KEY (id);


--
-- Name: build_summary build_summary_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_summary
    ADD CONSTRAINT build_summary_pkey PRIMARY KEY (id);


--
-- Name: build_failure_reason build_failure_reason_build_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_failure_reason
    ADD CONSTRAINT build_failure_reason_build_id_fkey FOREIGN KEY (build_id) REFERENCES build_summary(id);


--
-- Name: build_output build_output_build_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY build_output
    ADD CONSTRAINT build_output_build_id_fkey FOREIGN KEY (build_id) REFERENCES build_summary(id);


--
-- PostgreSQL database dump complete
--

