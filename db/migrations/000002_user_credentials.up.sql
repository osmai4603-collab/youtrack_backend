-- إضافة عمود كلمة المرور المشفّرة لجدول المستخدمين لدعم المصادقة (bcrypt).
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;
