-- Comments table
CREATE TABLE IF NOT EXISTS comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id   UUID REFERENCES comments(id) ON DELETE CASCADE,
    thread_id   UUID NOT NULL,
    user_id     TEXT NOT NULL,
    content     TEXT NOT NULL,
    reply_count INT DEFAULT 0,
    upvotes     INT DEFAULT 0,
    downvotes   INT DEFAULT 0,
    likes       INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT current_timestamp
);

-- Indexes for efficient sorting
CREATE INDEX IF NOT EXISTS idx_comments_thread_created ON comments(thread_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_thread_replies ON comments(thread_id, reply_count DESC);
CREATE INDEX IF NOT EXISTS idx_comments_thread_upvotes ON comments(thread_id, upvotes DESC);

-- Reactions table
CREATE TABLE IF NOT EXISTS comment_reactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id  UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('like', 'upvote', 'downvote')),
    created_at  TIMESTAMPTZ DEFAULT current_timestamp,

    UNIQUE (comment_id, user_id, type)
);

-- Index to quickly fetch reactions per comment
CREATE INDEX IF NOT EXISTS idx_comment_reactions_comment ON comment_reactions(comment_id);
