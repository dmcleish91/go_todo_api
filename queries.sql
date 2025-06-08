CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE todos (
    todo_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_completed BOOLEAN DEFAULT FALSE,
    due_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE tags (
    tag_id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    CONSTRAINT unique_tag_name UNIQUE (name)
);

CREATE TABLE refresh_tokens (
    token_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add index for faster token lookups
CREATE INDEX idx_refresh_token ON refresh_tokens(token);

-- Signup: Insert a new user
INSERT INTO users (username, email, password_hash) 
VALUES ($1, $2, $3);

-- Login: Fetch user by email to verify credentials
SELECT user_id, username, password_hash 
FROM users 
WHERE email = $1;

-- Get list of todos for a specific user
SELECT 
    todo_id, 
    title, 
    description, 
    is_completed, 
    due_date, 
    created_at, 
    updated_at,
    tags
FROM todos
WHERE user_id = $1
ORDER BY created_at DESC;

-- Add a new todo
INSERT INTO todos (user_id, title, description, due_date, tags) 
VALUES ($1, $2, $3, $4, $5);

-- Edit an existing todo
UPDATE todos 
SET title = $3,
    description = $4,
    due_date = $5,
    tags = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE todo_id = $1 
AND user_id = $2
RETURNING todo_id, title, description, due_date, is_completed, updated_at, tags;

-- Mark a todo as complete
UPDATE todos 
SET is_completed = TRUE, updated_at = CURRENT_TIMESTAMP 
WHERE todo_id = $1 AND user_id = $2;

-- Delete a todo
DELETE FROM todos 
WHERE todo_id = $1 AND user_id = $2;

-- Add a new tag preset (if needed)
INSERT INTO tags (name) 
VALUES ($1) 
RETURNING tag_id;
