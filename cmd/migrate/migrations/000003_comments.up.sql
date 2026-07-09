CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    post_id bigserial NOT NULL,
    user_id bigserial NOT NULL,
    content text NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
