-- UUIDを生成するための拡張機能を有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- users テーブル
CREATE TABLE users (
    uid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth0_sub VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- partners テーブル
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(36) NOT NULL,
    personality VARCHAR(12) NOT NULL,
    hair_color VARCHAR(12) NOT NULL,
    voice_type VARCHAR(36) NOT NULL,
    current_stage VARCHAR(50) NOT NULL,
    stamina INTEGER NOT NULL,
    intelligence INTEGER NOT NULL,
    sense INTEGER NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(uid)
);

-- scenarios テーブル
CREATE TABLE scenarios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stage VARCHAR(36) NOT NULL,
    routes CHAR(36) NOT NULL,
    template_text TEXT NOT NULL,
    stat_effect JSONB NOT NULL,
    weight INTEGER NOT NULL
);

-- sleep_logs テーブル
CREATE TABLE sleep_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    partner_id UUID NOT NULL,
    slept_at TIMESTAMP NOT NULL,
    wake_at TIMESTAMP,
    sleep_minutes INTEGER,
    CONSTRAINT fk_log_user FOREIGN KEY(user_id) REFERENCES users(uid),
    CONSTRAINT fk_log_partner FOREIGN KEY(partner_id) REFERENCES partners(id)
);

-- memories テーブル
CREATE TABLE memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_id UUID NOT NULL,
    scenario_id UUID NOT NULL,
    generated_prompt TEXT NOT NULL,
    occurred_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_memory_partner FOREIGN KEY(partner_id) REFERENCES partners(id),
    CONSTRAINT fk_memory_scenario FOREIGN KEY(scenario_id) REFERENCES scenarios(id)
);

-- partner_images テーブル
CREATE TABLE partner_images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_id UUID NOT NULL,
    stage VARCHAR(20) NOT NULL,
    image_url TEXT,
    generation_prompt TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_image_partner FOREIGN KEY(partner_id) REFERENCES partners(id)
);