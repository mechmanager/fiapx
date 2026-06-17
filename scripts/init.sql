-- Script de inicialização do banco de dados PostgreSQL do FIAP X.
-- Executado automaticamente pelo container postgres na primeira subida
-- (montado em /docker-entrypoint-initdb.d via docker-compose).

-- Habilita a extensão pgcrypto para usar gen_random_uuid().
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Tabela de usuários do sistema.
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- identificador único do usuário
    name          VARCHAR(120) NOT NULL,                       -- nome do usuário
    email         VARCHAR(255) UNIQUE NOT NULL,                -- email único usado no login
    password_hash VARCHAR(255) NOT NULL,                       -- hash bcrypt da senha
    created_at    TIMESTAMP DEFAULT NOW()                      -- data de criação do registro
);

-- Tabela de vídeos enviados e seu status de processamento.
CREATE TABLE IF NOT EXISTS videos (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),        -- identificador único do vídeo
    user_id           UUID NOT NULL REFERENCES users(id),                -- dono do vídeo
    original_filename VARCHAR(255) NOT NULL,                             -- nome original do arquivo enviado
    s3_key            VARCHAR(500) NOT NULL,                             -- chave do vídeo no MinIO
    status            VARCHAR(20) NOT NULL DEFAULT 'PENDING',            -- PENDING, PROCESSING, DONE ou ERROR
    error_message     TEXT,                                             -- mensagem de erro, quando status = ERROR
    zip_s3_key        VARCHAR(500),                                     -- chave do zip de frames no MinIO
    frame_count       INT DEFAULT 0,                                    -- quantidade de frames extraídos
    created_at        TIMESTAMP DEFAULT NOW(),                          -- data de criação do registro
    updated_at        TIMESTAMP DEFAULT NOW()                           -- data da última atualização
);

-- Índice para acelerar a listagem de vídeos por usuário.
CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id);

-- Garante que o status só assuma valores válidos.
ALTER TABLE videos DROP CONSTRAINT IF EXISTS chk_videos_status;
ALTER TABLE videos ADD CONSTRAINT chk_videos_status
    CHECK (status IN ('PENDING', 'PROCESSING', 'DONE', 'ERROR'));
