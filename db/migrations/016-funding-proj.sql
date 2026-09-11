-- Which venture a funding entry serves: palantir (Veridical Earth / "Palantir for
-- land use", YC application), ngi (Landscape Governance Initiative, CAR) or both.
ALTER TABLE funding ADD COLUMN proj TEXT NOT NULL DEFAULT 'palantir';

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (016, '016-funding-proj');
