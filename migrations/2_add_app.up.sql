INSERT INTO apps (id, name, secret)
VALUES (1, 'krypsy', 'test-secret')
ON CONFLICT DO NOTHING;