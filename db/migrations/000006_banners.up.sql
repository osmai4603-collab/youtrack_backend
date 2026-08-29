-- 000006_banners.up.sql
-- New independent schema for the banners block of GET /api/config
-- (request11.txt: fields=banners(globalBanner,globalBannerEnabled,systemEventsBanners)).
-- No existing table is modified.

CREATE TABLE banners_config (
    id SERIAL PRIMARY KEY,
    global_banner TEXT NOT NULL DEFAULT '',
    global_banner_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE system_events_banners (
    id SERIAL PRIMARY KEY,
    config_id INT REFERENCES banners_config(id) ON DELETE CASCADE,
    text TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_system_events_banners_config ON system_events_banners(config_id);