ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS channel VARCHAR(20) NOT NULL DEFAULT 'email';

ALTER TABLE notifications
    DROP CONSTRAINT IF EXISTS notifications_outbox_id_fkey;

INSERT INTO permissions (name, description)
VALUES ('review:moderate', 'menghapus ulasan orang lain')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description)
VALUES ('queue:monitor', 'memantau antrean pekerjaan latar belakang')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name)
VALUES ('admin', 'review:moderate')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name)
VALUES ('admin', 'queue:monitor')
ON CONFLICT DO NOTHING;
