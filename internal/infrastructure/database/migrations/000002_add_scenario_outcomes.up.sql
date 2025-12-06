-- scenarios テーブルに判定と結果用のカラムを追加
ALTER TABLE scenarios
ADD COLUMN condition_stat VARCHAR(50),
ADD COLUMN condition_value INTEGER DEFAULT 0,
ADD COLUMN success_text TEXT,
ADD COLUMN failure_text TEXT,
ADD COLUMN success_effect JSONB DEFAULT '{}' NOT NULL,
ADD COLUMN failure_effect JSONB DEFAULT '{}' NOT NULL;

-- 既存のデータを壊さないためのコメント（後で役立ちます）
COMMENT ON COLUMN scenarios.condition_stat IS '判定対象のステータス (例: stamina)';
COMMENT ON COLUMN scenarios.condition_value IS '判定の閾値';