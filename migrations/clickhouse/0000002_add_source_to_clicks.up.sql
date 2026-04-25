ALTER TABLE clicks ADD COLUMN source LowCardinality(String) DEFAULT 'rest';
ALTER TABLE clicks ADD COLUMN city    LowCardinality(String);
ALTER TABLE clicks ADD COLUMN asn     String;