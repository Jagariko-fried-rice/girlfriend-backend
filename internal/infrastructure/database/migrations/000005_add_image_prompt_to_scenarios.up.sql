-- シナリオテーブルに画像生成用プロンプトのカラムを追加
ALTER TABLE scenarios ADD COLUMN image_prompt TEXT;

-- 既存のデータは一旦、日本語のテキストをコピーしておく（NULL回避）
UPDATE scenarios SET image_prompt = template_text;