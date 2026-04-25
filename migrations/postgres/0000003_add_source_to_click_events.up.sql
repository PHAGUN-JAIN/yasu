ALTER TABLE click_events ADD COLUMN source VARCHAR(20) DEFAULT 'rest';
ALTER TABLE click_events ADD COLUMN city    VARCHAR(100);
ALTER TABLE click_events ADD COLUMN asn     VARCHAR(255);