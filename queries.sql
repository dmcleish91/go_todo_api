CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

create table projects (
  project_id uuid not null default gen_random_uuid (),
  user_id uuid not null,
  project_name character varying not null,
  color character varying null,
  is_inbox boolean null default false,
  parent_project_id uuid null,
  constraint projects_pkey primary key (project_id),
  constraint projects_parent_project_id_fkey foreign KEY (parent_project_id) references projects (project_id),
  constraint projects_user_id_fkey foreign KEY (user_id) references auth.users (id)
)

create table tasks (
  task_id uuid not null default gen_random_uuid (),
  project_id uuid not null,
  user_id uuid not null,
  content text not null,
  description text null,
  due_date date null,
  date_datetime time with time zone null,
  priority smallint null,
  is_completed boolean null default false,
  completed_at timestamp with time zone null,
  parent_task_id uuid null,
  constraint Tasks_pkey primary key (task_id),
  constraint tasks_parent_task_id_fkey foreign KEY (parent_task_id) references tasks (task_id) on update CASCADE on delete CASCADE,
  constraint tasks_project_id_fkey foreign KEY (project_id) references projects (project_id),
  constraint tasks_user_id_fkey foreign KEY (user_id) references auth.users (id)
)

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
