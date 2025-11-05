CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Добавляем проверку что end_date либо NULL, либо больше start_date
    CONSTRAINT chk_dates_valid CHECK (end_date IS NULL OR end_date > start_date)
);

-- Обновляем триггер чтобы он также проверял даты
CREATE OR REPLACE FUNCTION set_default_end_date()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.end_date IS NULL THEN
        NEW.end_date := NEW.start_date + INTERVAL '1 month';
    ELSIF NEW.end_date <= NEW.start_date THEN
        RAISE EXCEPTION 'end_date must be greater than start_date';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_end_date_before_insert
    BEFORE INSERT ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION set_default_end_date();

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);
CREATE INDEX idx_subscriptions_user_service ON subscriptions (user_id, service_name);