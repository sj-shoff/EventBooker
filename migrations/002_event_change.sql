-- +goose Up
-- +goose StatementBegin

-- Добавляем поле status в таблицу events
ALTER TABLE events 
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active',
ADD CONSTRAINT events_status_check 
CHECK (status IN ('active', 'cancelled', 'completed'));

-- Обновляем существующие записи
UPDATE events SET status = 'active' WHERE status IS NULL OR status = '';

-- Создаем индекс для быстрого поиска по статусу
CREATE INDEX idx_events_status ON events(status);

-- Обновляем представления или функции, если они есть
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем индекс
DROP INDEX IF EXISTS idx_events_status;

-- Удаляем ограничение
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_status_check;

-- Удаляем столбец status
ALTER TABLE events DROP COLUMN IF EXISTS status;

-- +goose StatementEnd