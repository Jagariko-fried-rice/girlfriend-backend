-- routes カラムを CHAR(36) から TEXT に変更する
-- これにより、固定長パディング（空白埋め）による検索ミスを防ぎます
ALTER TABLE scenarios ALTER COLUMN routes TYPE TEXT;