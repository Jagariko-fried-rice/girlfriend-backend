-- すべてのテーブルでRLS（行レベルセキュリティ）を有効化
-- これにより、明示的なポリシーがない限り、外部(PostgREST)からのアクセスはすべて遮断されます。
-- Goのバックエンド(postgresユーザー)はBYPASSRLS権限を持つため、影響を受けずにアクセス可能です。

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE partners ENABLE ROW LEVEL SECURITY;
ALTER TABLE scenarios ENABLE ROW LEVEL SECURITY;
ALTER TABLE sleep_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE memories ENABLE ROW LEVEL SECURITY;
ALTER TABLE partner_images ENABLE ROW LEVEL SECURITY;

-- 管理用テーブルも警告対象になるため有効化しておきます
ALTER TABLE schema_migrations ENABLE ROW LEVEL SECURITY;