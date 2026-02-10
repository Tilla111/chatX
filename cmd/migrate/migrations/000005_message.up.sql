CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE messages (
   id          BIGSERIAL PRIMARY KEY, 
   chat_id     BIGINT NOT NULL REFERENCES chats(id) ON DELETE CASCADE, 
   sender_id   BIGINT REFERENCES users(id) ON DELETE SET NULL, 
   message_text TEXT NOT NULL, 
   created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
   updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_messages_modtime
BEFORE UPDATE ON messages
FOR EACH ROW
EXECUTE PROCEDURE update_modified_column();