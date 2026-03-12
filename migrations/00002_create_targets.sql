-- +goose Up
CREATE TABLE targets(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    URL TEXT NOT NULL,
    interval_time INTEGER NOT NULL,
    timeout INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX inx_targets_user_id ON targets(user_id);
-- +goose Down
DROP TABLE targets;