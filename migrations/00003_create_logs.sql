-- +goose Up
CREATE TABLE check_logs
(
    id         BIGSERIAL PRIMARY KEY,
    status        TEXT   NOT NULL,
    status_code INT ,
    error_msg TEXT ,
    response_time_ms INT ,
    target_id  BIGINT NOT NULL,
    checked_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE
);
CREATE INDEX idx_logs_target_id ON check_logs(target_id);

-- +goose Down
DROP TABLE check_logs;
