UPDATE users SET password_hash = '$2a$10$YewTN5skmEf3lKyp/GK6V.V4UZ4eeT3aaFsemYhILNWupR6LXz4bq' WHERE username = 'admin';
SELECT username, password_hash FROM users WHERE username = 'admin';
