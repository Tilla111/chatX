CREATE TABLE group_info (
  id BIGSERIAL PRIMARY KEY,
  chat_id BIGINT UNIQUE REFERENCES chats(id) ON DELETE CASCADE,
  group_name VARCHAR(100),
  group_description TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_group_info_modtime
BEFORE UPDATE ON group_info
FOR EACH ROW
EXECUTE PROCEDURE update_modified_column();