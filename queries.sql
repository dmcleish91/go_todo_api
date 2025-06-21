-- The queries.sql file is a reference for all the SQL statements used in the application.

-- The queries below are used to create the tables in the database.
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
);

create table public.tasks (
  task_id uuid not null default gen_random_uuid (),
  project_id uuid null,
  user_id uuid not null,
  content text not null,
  description text null,
  due_date date null,
  due_datetime time without time zone null,
  priority smallint null,
  is_completed boolean null default false,
  completed_at timestamp with time zone null,
  parent_task_id uuid null,
  created_at timestamp with time zone not null default now(),
  constraint Tasks_pkey primary key (task_id),
  constraint tasks_parent_task_id_fkey foreign KEY (parent_task_id) references tasks (task_id) on update CASCADE on delete CASCADE,
  constraint tasks_project_id_fkey foreign KEY (project_id) references projects (project_id),
  constraint tasks_user_id_fkey foreign KEY (user_id) references auth.users (id)
);


-- The queries below are used in the projects model.

-- AddProject
INSERT INTO projects (
    user_id, project_name, color, is_inbox, parent_project_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING project_id, user_id, project_name, color, is_inbox, parent_project_id;

-- EditProjectByID
UPDATE projects SET
    project_name = $3,
    color = $4,
    is_inbox = $5,
    parent_project_id = $6
WHERE project_id = $1 AND user_id = $2
RETURNING project_id, user_id, project_name, color, is_inbox, parent_project_id;

-- GetProjectsByUserID
SELECT project_id, user_id, project_name, color, is_inbox, parent_project_id
FROM projects
WHERE user_id = $1
ORDER BY project_name ASC;

-- DeleteProjectByID
DELETE FROM projects WHERE project_id = $1 AND user_id = $2;


-- The queries below are used in the tasks model.

-- AddTask
INSERT INTO tasks (
    project_id, user_id, content, description, due_date, due_datetime, priority, parent_task_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id;

-- EditTaskByID
UPDATE tasks SET
    project_id = $3,
    content = $4,
    description = $5,
    due_date = $6,
    due_datetime = $7,
    priority = $8,
    is_completed = $9,
    completed_at = $10,
    parent_task_id = $11
WHERE task_id = $1 AND user_id = $2
RETURNING task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id;

-- GetTasksByUserID
SELECT task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, created_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at ASC;

-- ToggleTaskCompleted
UPDATE tasks SET is_completed = NOT is_completed, completed_at = CASE WHEN is_completed THEN NULL ELSE CURRENT_TIMESTAMP END WHERE task_id = $1 AND user_id = $2;

-- DeleteTaskByID
DELETE FROM tasks WHERE task_id = $1 AND user_id = $2;
