-- สร้างตาราง users สำหรับระบบ login
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `username` VARCHAR(50) NOT NULL UNIQUE,
    `email` VARCHAR(100) NOT NULL UNIQUE,
    `password_hash` VARCHAR(255) NOT NULL,
    `full_name` VARCHAR(100),
    `role` ENUM('admin', 'editor', 'viewer') DEFAULT 'viewer',
    `is_active` BOOLEAN DEFAULT TRUE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `last_login` TIMESTAMP NULL
);

-- สร้าง sessions table สำหรับจัดการ session
CREATE TABLE IF NOT EXISTS `sessions` (
    `id` VARCHAR(128) PRIMARY KEY,
    `user_id` INT NOT NULL,
    `expires_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
);

-- เพิ่มข้อมูลผู้ใช้เริ่มต้น (admin)
-- รหัสผ่าน: admin123 (hashed)
INSERT INTO `users` (`username`, `email`, `password_hash`, `full_name`, `role`) 
VALUES ('admin', 'admin@ballthai.com', '$2a$10$YewTN5skmEf3lKyp/GK6V.V4UZ4eeT3aaFsemYhILNWupR6LXz4bq', 'ผู้ดูแลระบบ', 'admin')
ON DUPLICATE KEY UPDATE `username` = `username`;
