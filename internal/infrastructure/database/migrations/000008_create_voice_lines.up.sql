CREATE TABLE voice_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    personality VARCHAR(20) NOT NULL, -- 元気, おっとり, etc.
    situation VARCHAR(20) NOT NULL,   -- morning, night, home
    line_text TEXT NOT NULL,          -- セリフ本文
    audio_url TEXT,                   -- 将来用: 音声ファイルのURL
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 検索を速くするためのインデックス
CREATE INDEX idx_voice_lines_personality ON voice_lines(personality);