CREATE TABLE IF NOT EXISTS communications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier TEXT NOT NULL UNIQUE, -- e.g. "+18045551234" or "user@example.com"
    type TEXT NOT NULL CHECK (type IN ('phone', 'email'))
);

CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS conversation_memberships (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    communication_id UUID NOT NULL REFERENCES communications(id) ON DELETE CASCADE,
    PRIMARY KEY (conversation_id, communication_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    sender_id UUID NOT NULL REFERENCES communications(id),
    provider_id TEXT, -- external message ID
    message_type TEXT NOT NULL CHECK (message_type IN ('mms', 'sms', 'email')),
    body TEXT,
    attachments TEXT[], 
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL
);
