-- migrate:up
CREATE
    EXTENSION IF NOT EXISTS pgcrypto;

-- Function for deployment notifications.
CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS
$$
DECLARE
    topic TEXT;
    payload JSONB;
BEGIN
   topic := CASE TG_TABLE_NAME
        WHEN 'deployments' THEN 'deployments_events'
        WHEN 'topics' THEN 'topics_events'
        WHEN 'topic_events' THEN 'topic_events_events'
    END;
    IF TG_OP = 'DELETE'
    THEN
        payload = jsonb_build_object(
                'table', TG_TABLE_NAME,
                'action', TG_OP,
                'old', old.key
            );
    ELSE
        payload = jsonb_build_object(
                'table', TG_TABLE_NAME,
                'action', TG_OP,
                'new', new.key
            );
    END IF;
    PERFORM pg_notify(topic, payload::text);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE modules
(
    id       BIGINT         NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    language TEXT        NOT NULL,
    name     TEXT UNIQUE NOT NULL
);

-- [<module>.]<name> represented as a schema.Ref
CREATE DOMAIN schema_ref AS TEXT;

-- The parseable string representation of a schema.Type.
CREATE DOMAIN schema_type AS BYTEA;

-- Proto-encoded module schema.
CREATE DOMAIN module_schema_pb AS BYTEA;

CREATE DOMAIN deployment_key AS TEXT;

CREATE TABLE deployments
(
    id           BIGINT         NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    module_id    BIGINT         NOT NULL REFERENCES modules (id) ON DELETE CASCADE,
    -- Unique key for this deployment in the form <module-name>-<random>.
    "key"       deployment_key UNIQUE NOT NULL,
    "schema"     module_schema_pb  NOT NULL,
    -- Labels are used to match deployments to runners.
    "labels"     JSONB          NOT NULL DEFAULT '{}',
    min_replicas INT            NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX deployments_key_idx ON deployments (key);
CREATE INDEX deployments_module_id_idx ON deployments (module_id);
-- Only allow one deployment per module.
CREATE UNIQUE INDEX deployments_unique_idx ON deployments (module_id)
    WHERE min_replicas > 0;

CREATE TRIGGER deployments_notify_event
    AFTER INSERT OR UPDATE OR DELETE
    ON deployments
    FOR EACH ROW
EXECUTE PROCEDURE notify_event();

CREATE TABLE artefacts
(
    id         BIGINT       NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    -- SHA256 digest of the content.
    digest     BYTEA UNIQUE NOT NULL,
    content    BYTEA        NOT NULL
);

CREATE UNIQUE INDEX artefacts_digest_idx ON artefacts (digest);

CREATE TABLE deployment_artefacts
(
    artefact_id   BIGINT      NOT NULL REFERENCES artefacts (id) ON DELETE CASCADE,
    deployment_id BIGINT      NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    executable    BOOLEAN     NOT NULL,
    -- Path relative to the module root.
    path          TEXT     NOT NULL
);

CREATE INDEX deployment_artefacts_deployment_id_idx ON deployment_artefacts (deployment_id);

CREATE TYPE runner_state AS ENUM (
    -- The Runner is available to run deployments.
    'idle',
    -- The Runner is reserved but has not yet deployed.
    'reserved',
    -- The Runner has been assigned a deployment.
    'assigned',
    -- The Runner is dead.
    'dead'
    );

CREATE DOMAIN runner_key AS TEXT;

-- Runners are processes that are available to run modules.
CREATE TABLE runners
(
    id                  BIGINT       NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    -- Unique identifier for this runner, generated at startup.
    key                 runner_key UNIQUE  NOT NULL,
    created             TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    last_seen           TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    -- If the runner is reserved, this is the time at which the reservation expires.
    reservation_timeout TIMESTAMPTZ,
    state               runner_state NOT NULL DEFAULT 'idle',
    endpoint            TEXT      NOT NULL,
    -- Some denormalisation for performance. Without this we need to do a two table join.
    module_name         TEXT,
    deployment_id       BIGINT       REFERENCES deployments (id) ON DELETE SET NULL,
    labels              JSONB        NOT NULL DEFAULT '{}'
);

-- Automatically update module_name when deployment_id is set or unset.
CREATE OR REPLACE FUNCTION runners_update_module_name() RETURNS TRIGGER AS
$$
BEGIN
    IF NEW.deployment_id IS NULL
    THEN
        NEW.module_name = NULL;
    ELSE
        SELECT m.name
        INTO NEW.module_name
        FROM modules m
                 INNER JOIN deployments d on m.id = d.module_id
        WHERE d.id = NEW.deployment_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER runners_update_module_name
    BEFORE INSERT OR UPDATE
    ON runners
    FOR EACH ROW
EXECUTE PROCEDURE runners_update_module_name();

-- Set a default reservation_timeout when a runner is reserved.
CREATE OR REPLACE FUNCTION runners_set_reservation_timeout() RETURNS TRIGGER AS
$$
BEGIN
    IF OLD.state != 'reserved' AND NEW.state = 'reserved' AND NEW.reservation_timeout IS NULL
    THEN
        NEW.reservation_timeout = NOW() AT TIME ZONE 'utc' + INTERVAL '2 minutes';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER runners_set_reservation_timeout
    BEFORE INSERT OR UPDATE
    ON runners
    FOR EACH ROW
EXECUTE PROCEDURE runners_set_reservation_timeout();

CREATE UNIQUE INDEX runners_key ON runners (key);
CREATE UNIQUE INDEX runners_endpoint_not_dead_idx ON runners (endpoint) WHERE state <> 'dead';
CREATE INDEX runners_module_name_idx ON runners (module_name);
CREATE INDEX runners_state_idx ON runners (state);
CREATE INDEX runners_deployment_id_idx ON runners (deployment_id);
CREATE INDEX runners_labels_idx ON runners USING GIN (labels);

CREATE TABLE ingress_routes
(
    method        TEXT NOT NULL,
    path          TEXT NOT NULL,
    -- The deployment that should handle this route.
    deployment_id BIGINT  NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    -- Duplicated here to avoid having to join from this to deployments then modules.
    module        TEXT NOT NULL,
    verb          TEXT NOT NULL
);

CREATE INDEX ingress_routes_method_path_idx ON ingress_routes (method, path);

CREATE TYPE origin AS ENUM (
    'ingress',
    'cron',
    -- Not supported yet.
    'pubsub'
    );

CREATE DOMAIN request_key AS TEXT;

-- Requests originating from outside modules, either from external sources or from
-- events within the system.
CREATE TABLE requests
(
    id          BIGINT         NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    origin      origin         NOT NULL,
    -- Will be in the form <origin>-<description>-<hash>:
    --
    -- ingress: ingress-<method>-<path>-<hash> (eg. ingress-GET-foo-bar-<hash>)
    -- cron: cron-<name>-<hash>                (eg. cron-poll-news-sources-<hash>)
    -- pubsub: pubsub-<subscription>-<hash>    (eg. pubsub-articles-<hash>)
    "key"       request_key UNIQUE NOT NULL,
    source_addr TEXT        NOT NULL
);

CREATE INDEX requests_origin_idx ON requests (origin);
CREATE UNIQUE INDEX ingress_requests_key_idx ON requests ("key");

CREATE TYPE controller_state AS ENUM (
    'live',
    'dead'
    );

CREATE DOMAIN controller_key AS TEXT;

CREATE TABLE controller
(
    id        BIGINT           NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    key       controller_key UNIQUE      NOT NULL,
    created   TIMESTAMPTZ      NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    last_seen TIMESTAMPTZ      NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    state     controller_state NOT NULL DEFAULT 'live',
    endpoint  TEXT          NOT NULL
);

CREATE UNIQUE INDEX controller_endpoint_not_dead_idx ON controller (endpoint) WHERE state <> 'dead';

CREATE TYPE cron_job_state AS ENUM (
    'idle',
    'executing'
);

CREATE DOMAIN cron_job_key AS TEXT;

CREATE TABLE cron_jobs
(
    id             BIGINT      NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    key            cron_job_key UNIQUE NOT NULL,
    deployment_id  BIGINT      NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    verb           TEXT     NOT NULL,
    schedule       TEXT	   NOT NULL,
    start_time     TIMESTAMPTZ NOT NULL,
    next_execution TIMESTAMPTZ NOT NULL,
    state          cron_job_state   NOT NULL DEFAULT 'idle',

    -- Some denormalisation for performance. Without this we need to do a two table join.
    module_name    TEXT     NOT NULL
);

CREATE INDEX cron_jobs_executing_start_time_idx ON cron_jobs (start_time) WHERE state = 'executing';
CREATE UNIQUE INDEX cron_jobs_key_idx ON cron_jobs (key);

CREATE TYPE event_type AS ENUM (
    'call',
    'log',
    'deployment_created',
    'deployment_updated'
);

CREATE TABLE events
(
    id            BIGINT      NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    time_stamp    TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),

    deployment_id BIGINT      NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
    request_id    BIGINT      NULL REFERENCES requests (id) ON DELETE CASCADE,

    type          event_type  NOT NULL,

    -- Type-specific keys used to index events for searching.
    custom_key_1  TEXT     NULL,
    custom_key_2  TEXT     NULL,
    custom_key_3  TEXT     NULL,
    custom_key_4  TEXT     NULL,

    payload       JSONB       NOT NULL
);

CREATE INDEX events_timestamp_idx ON events (time_stamp);
CREATE INDEX events_deployment_id_idx ON events (deployment_id);
CREATE INDEX events_request_id_idx ON events (request_id);
CREATE INDEX events_type_idx ON events (type);
CREATE INDEX events_custom_key_1_idx ON events (custom_key_1);
CREATE INDEX events_custom_key_2_idx ON events (custom_key_2);
CREATE INDEX events_custom_key_3_idx ON events (custom_key_3);
CREATE INDEX events_custom_key_4_idx ON events (custom_key_4);

CREATE DOMAIN topic_key AS TEXT;

-- Topics are a way to asynchronously publish data between modules.
CREATE TABLE topics (
    id BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    "key" topic_key UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),

    -- Each topic is associated with an owning module.
    module_id BIGINT NOT NULL REFERENCES modules(id),

    -- Name of the topic.
    name TEXT NOT NULL,

    -- Data reference to the payload data type in the owning module's schema.
    type TEXT NOT NULL
);

CREATE UNIQUE INDEX topics_module_name_idx ON topics(module_id, name);

CREATE TRIGGER topics_notify_event
    AFTER INSERT OR UPDATE OR DELETE
    ON topics
    FOR EACH ROW
EXECUTE PROCEDURE notify_event();

-- This table contains the actual topic data.
CREATE TABLE topic_events (
    id BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),

    topic_id BIGINT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,

    payload BYTEA NOT NULL
);

CREATE TRIGGER topic_events_notify_event
    AFTER INSERT OR UPDATE OR DELETE
    ON topic_events
    FOR EACH ROW
EXECUTE PROCEDURE notify_event();

CREATE DOMAIN subscription_key AS TEXT;

-- A subscription to a topic.
--
-- Multiple subscribers can consume from a single subscription.
CREATE TABLE topic_subscriptions (
    id BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    "key" subscription_key UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),

    topic_id BIGINT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,

     -- Each subscription is associated with an owning module.
    module_id BIGINT NOT NULL REFERENCES modules(id),

    -- Name of the subscription.
    name TEXT UNIQUE NOT NULL,

    -- Cursor pointing into the topic_events table.
    cursor BIGINT REFERENCES topic_events(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX topic_subscriptions_module_name_idx ON topic_subscriptions(module_id, name);

CREATE DOMAIN subscriber_key AS TEXT;

-- A subscriber to a topic.
--
-- A subscriber is a 1:1 mapping between a subscription and a sink.
CREATE TABLE topic_subscribers (
   id BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
   "key" subscriber_key UNIQUE NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),

   topic_subscriptions_id BIGINT NOT NULL REFERENCES topic_subscriptions(id),

   deployment_id BIGINT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
   -- Name of the verb to call on the deployment.
   sink TEXT NOT NULL
);

CREATE DOMAIN lease_key AS TEXT;

CREATE TABLE leases (
    id BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    idempotency_key UUID UNIQUE NOT NULL,
    key lease_key UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc')
);

CREATE INDEX leases_expires_at_idx ON leases (expires_at);

CREATE TYPE async_call_state AS ENUM (
    'pending', -- The call is scheduled to be executed.
    'executing',  -- A controller is executing the call.
    'success', -- The call was successful.
    'error' -- The call failed and "error" is populated.
);

-- An asynchronous call to a verb.
CREATE TABLE async_calls (
    id BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),

    -- Lease obtained when the call is executed. Will be set to NULL if the
    -- lease expires (ie. controller died). Calls in this state can be selected
    -- with (lease_id IS NULL AND async_call_state != 'pending')
    lease_id BIGINT REFERENCES leases(id) ON DELETE SET NULL,

    verb schema_ref NOT NULL,
    state async_call_state NOT NULL DEFAULT 'pending',
    -- Originator of the call (cron, fsm, pubsub) in the form <type>:<payload>. See AsyncCallOrigin sum type.
    origin TEXT NOT NULL,
    -- earliest timestamp to execute the call
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    -- Request to send to the verb.
    request JSONB NOT NULL,
    -- Populated on success.
    response JSONB,
    -- Populated on error.
    error TEXT,

    -- retry state
    remaining_attempts INT NOT NULL,
    backoff            INTERVAL NOT NULL,
    max_backoff        INTERVAL NOT NULL
);

CREATE INDEX async_calls_state_idx ON async_calls (state);

CREATE TYPE fsm_status AS ENUM ('running', 'completed', 'failed');

CREATE TABLE fsm_instances (
    id BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    -- Reference to the FSM schema.
    fsm schema_ref NOT NULL,
    -- Unique user-provided key identifying the instance.
    -- In the example from the design doc this would be the "invoice ID".
    key TEXT NOT NULL,
    status fsm_status NOT NULL DEFAULT 'running'::fsm_status,
    -- The current state of the FSM. NULL indicates an origin or terminal state.
    current_state schema_ref,
    -- Destination state for the active transition or NULL if the FSM is idle.
    destination_state schema_ref,
    -- Call handling the current transition. Will be NULL if the FSM is idle.
    async_call_id BIGINT REFERENCES async_calls(id)
);

CREATE UNIQUE INDEX idx_fsm_instances_fsm_key ON fsm_instances(fsm, key);
CREATE INDEX idx_fsm_instances_status ON fsm_instances(status);

CREATE TABLE module_configuration
(
    id         BIGINT NOT NULL GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    module     TEXT,  -- If NULL, configuration is global.
    name       TEXT   NOT NULL,
    value      JSONB  NOT NULL,
    UNIQUE (module, name)
);

-- migrate:down
