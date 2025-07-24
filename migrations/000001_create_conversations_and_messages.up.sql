CREATE TYPE message_type AS ENUM ('mms', 'sms', 'email');
CREATE TYPE communication_type AS ENUM ('phone', 'email');

CREATE TABLE IF NOT EXISTS communications (
    id BIGSERIAL PRIMARY KEY,
    identifier TEXT NOT NULL UNIQUE, -- e.g. "+18045551234" or "user@example.com"
    communication_type communication_type NOT NULL
);

CREATE TABLE IF NOT EXISTS conversations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Speeds up ORDER BY on conversations.created_at
CREATE INDEX idx_conversations_created_at ON conversations(created_at DESC);

CREATE TABLE IF NOT EXISTS conversation_memberships (
    conversation_id BIGSERIAL NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    communication_id BIGSERIAL NOT NULL REFERENCES communications(id) ON DELETE CASCADE,
    PRIMARY KEY (conversation_id, communication_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGSERIAL NOT NULL REFERENCES conversations(id),
    sender_id BIGSERIAL NOT NULL REFERENCES communications(id),
    provider_id TEXT, -- external message ID
    message_type message_type NOT NULL,
    body TEXT,
    attachments TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Speeds up JOIN: conversations -> messages and ORDER BY messages.created_at
CREATE INDEX idx_messages_conversation_id_created_at ON messages(conversation_id, created_at);

-- Speeds up JOIN: messages -> communications BY messages.sender_id
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
