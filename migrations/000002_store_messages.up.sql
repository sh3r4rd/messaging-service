CREATE OR REPLACE FUNCTION create_message(
    p_from               text,
    p_to                 text,
    p_provider_id        text,
    p_message_type       text,
    p_communication_type text,
    p_body               text,
    p_attachments        text[],
    p_created_at          timestamptz
) RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
    v_from_id         bigint;
    v_to_id           bigint;
    v_conversation_id bigint;
    v_message_id      bigint;
BEGIN
    ------------------------------------------------------------------
    -- 1. Ensure both communications exist (upsert‑then‑select pattern)
    ------------------------------------------------------------------
    INSERT INTO communications (identifier, type)
    VALUES (p_from, p_communication_type)
    ON CONFLICT (identifier) DO NOTHING;

    INSERT INTO communications (identifier, type)
    VALUES (p_to, p_communication_type)
    ON CONFLICT (identifier) DO NOTHING;

    SELECT id  INTO v_from_id FROM communications WHERE identifier = p_from;
    SELECT id  INTO v_to_id   FROM communications WHERE identifier = p_to;

    ------------------------------------------------------------------
    -- 2. Locate an existing two‑party conversation (exact pair only)
    ------------------------------------------------------------------
    SELECT cm.conversation_id
      INTO v_conversation_id
      FROM conversation_memberships cm
     WHERE cm.communication_id IN (v_from_id, v_to_id)
     GROUP BY cm.conversation_id
     HAVING COUNT(*) = 2
        AND bool_and(cm.communication_id IN (v_from_id, v_to_id))
     LIMIT 1;

    ------------------------------------------------------------------
    -- 3. If none exists, create conversation and memberships
    ------------------------------------------------------------------
    IF v_conversation_id IS NULL THEN
        INSERT INTO conversations (created_at)
        VALUES (now())
        RETURNING id INTO v_conversation_id;

        INSERT INTO conversation_memberships (conversation_id, communication_id)
        VALUES
            (v_conversation_id, v_from_id),
            (v_conversation_id, v_to_id)
        ON CONFLICT DO NOTHING;
    END IF;

    ------------------------------------------------------------------
    -- 4. Insert the message
    ------------------------------------------------------------------
    INSERT INTO messages (
        conversation_id, 
        sender_id,
        provider_id,        
        message_type,
        body,               
        attachments,
        created_at
    )
    VALUES (
        v_conversation_id, 
        v_from_id,
        p_provider_id,      
        p_message_type,
        p_body,             
        p_attachments,
        p_created_at
    )
    RETURNING id INTO v_message_id;
    ------------------------------------------------------------------
    RETURN v_message_id;  -- handy if the caller needs it
END;
$$;
