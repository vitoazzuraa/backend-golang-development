-- Migrasi 010: notifikasi perubahan kuota (Tugas Mandiri D.2).
--
-- Satu event outbox course.quota_changed menghasilkan BANYAK notifikasi,
-- satu per peserta kelas, sehingga UNIQUE (outbox_id) tidak lagi cocok
-- sebagai idempotency key. Idempotency key diganti menjadi pasangan
-- (outbox_id, user_id): setiap mahasiswa paling banyak memiliki satu
-- notifikasi untuk setiap event outbox.
--
-- Cara mengirim SATU TASK PER MAHASISWA (relay membaca view di bawah):
--   1. Relay membaca quota_notification_targets untuk id outbox tersebut
--      dan menerbitkan satu task asynq per baris (satu per mahasiswa),
--      dengan task id "outbox-<id>-user-<user_id>" dan payload
--      NotificationJob yang membawa EnrollmentEvent dengan user_id mahasiswa
--      yang bersangkutan.
--   2. Setiap task hanya meng-claim SATU baris notifications
--      (outbox_id, user_id) lalu mengirim SATU email, keduanya di dalam
--      satu transaksi (claim + kirim se-atomik).
--   3. Bila email ke-7 dari 40 gagal, hanya task mahasiswa ke-7 yang
--      di-retry oleh Asynq; email ke-1 sampai ke-6 sudah ter-claim dan
--      ter-commit, sehingga retry TIDAK mengirim ulang email tersebut
--      (claim berikutnya bentrok dengan UNIQUE di bawah dan di-skip).

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_id_key;

DROP INDEX IF EXISTS notifications_outbox_unique;

CREATE UNIQUE INDEX IF NOT EXISTS notifications_outbox_user_unique
    ON notifications (outbox_id, user_id);

CREATE OR REPLACE VIEW quota_notification_targets AS
SELECT o.id AS outbox_id,
       e.user_id
FROM outbox o
JOIN enrollments e ON e.course_id = (o.payload->>'course_id')::int
WHERE o.event_type = 'course.quota_changed';
