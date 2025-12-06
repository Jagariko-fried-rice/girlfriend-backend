-- sleep_logs テーブルの日時カラムを、タイムゾーン付きに変更します
-- これにより、JST(+09:00)などの情報が正しく保存され、計算ズレがなくなります
ALTER TABLE sleep_logs ALTER COLUMN slept_at TYPE TIMESTAMP WITH TIME ZONE;
ALTER TABLE sleep_logs ALTER COLUMN wake_at TYPE TIMESTAMP WITH TIME ZONE;