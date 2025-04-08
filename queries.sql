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
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE tags (
    tag_id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    CONSTRAINT unique_tag_name UNIQUE (name)
);

CREATE TABLE todo_tags (
    todo_id INT NOT NULL,
    tag_id INT NOT NULL,
    PRIMARY KEY (todo_id, tag_id),
    FOREIGN KEY (todo_id) REFERENCES todos(todo_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(tag_id) ON DELETE CASCADE
);


-- Signup: Insert a new user
INSERT INTO users (username, email, password_hash) 
VALUES ($1, $2, $3);

-- Login: Fetch user by email to verify credentials
SELECT user_id, username, password_hash 
FROM users 
WHERE email = $1;

-- Get list of todos for a specific user
SELECT 
    t.todo_id, 
    t.title, 
    t.description, 
    t.is_completed, 
    t.due_date, 
    t.created_at, 
    t.updated_at, 
    STRING_AGG(tag.name, ', ') AS tags
FROM 
    todos t
LEFT JOIN 
    todo_tags tt ON t.todo_id = tt.todo_id
LEFT JOIN 
    tags tag ON tt.tag_id = tag.tag_id
WHERE 
    t.user_id = $1
GROUP BY 
    t.todo_id
ORDER BY 
    t.created_at DESC;

-- Add a new todo
INSERT INTO todos (user_id, title, description, due_date) 
VALUES ($1, $2, $3, $4);

-- Edit an existing todo
UPDATE todos 
SET title = $3,
    description = $4,
    due_date = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE todo_id = $1 
AND user_id = $2
RETURNING todo_id, title, description, due_date, is_completed, updated_at;

-- Mark a todo as complete
UPDATE todos 
SET is_completed = TRUE, updated_at = CURRENT_TIMESTAMP 
WHERE todo_id = $1 AND user_id = $2;

-- Delete a todo
DELETE FROM todos 
WHERE todo_id = $1 AND user_id = $2;

-- Add a tag to an existing todo
INSERT INTO todo_tags (todo_id, tag_id) 
VALUES ($1, $2);

-- Add a new tag (if needed)
INSERT INTO tags (name) 
VALUES ($1) 
RETURNING tag_id;
