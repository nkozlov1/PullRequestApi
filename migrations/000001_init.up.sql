CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY
    );

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(500) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(user_id)
    );

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (pull_request_id, user_id),
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
    );