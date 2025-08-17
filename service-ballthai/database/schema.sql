-- schema.sql
-- ไฟล์นี้รวบรวมคำสั่ง SQL สำหรับสร้างตารางฐานข้อมูลทั้งหมด

-- 1. สร้างตาราง leagues (ลีก)
-- ตารางนี้เก็บข้อมูลลีกต่างๆ
CREATE TABLE IF NOT EXISTS `leagues` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE
);

-- เพิ่มข้อมูลลีกตัวอย่าง
INSERT INTO leagues (id, name) VALUES
(1, 'ไทยลีก 1'),
(2, 'ไทยลีก 2'),
(3, 'ไทยลีก 3'),
(4, 'Revo League Cup'),
(5, 'FA Cup'),
(6, 'BGC Cup'),
(59, 'Samipro');

-- 2. สร้างตาราง stadiums (สนาม)
-- ตารางนี้เก็บข้อมูลสนามแข่งขัน
CREATE TABLE IF NOT EXISTS `stadiums` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `stadium_ref_id` INT UNIQUE,
    `team_id` INT,
    `name` VARCHAR(255) NOT NULL,
    `name_en` VARCHAR(255),
    `short_name` VARCHAR(255),
    `short_name_en` VARCHAR(255),
    `year_established` INT,
    `country_name` VARCHAR(255),
    `country_code` VARCHAR(10),
    `capacity` INT,
    `latitude` DOUBLE,
    `longitude` DOUBLE,
    `photo_url` VARCHAR(255)
);

-- 3. สร้างตาราง teams (ทีม)
-- ตารางนี้เก็บข้อมูลทีมฟุตบอล
CREATE TABLE IF NOT EXISTS `teams` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `name_th` VARCHAR(255) NOT NULL UNIQUE,
    `name_en` VARCHAR(255),
    `logo_url` VARCHAR(255),
    `team_post_ballthai` VARCHAR(255),
    `website` VARCHAR(255),
    `shop` VARCHAR(255),
    `stadium_id` INT -- Foreign Key จะถูกเพิ่มทีหลัง
);

-- 4. สร้างตาราง channels (ช่องทีวี)
-- ตารางนี้เก็บข้อมูลช่องทางการถ่ายทอดสด
CREATE TABLE IF NOT EXISTS `channels` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE,
    `logo_url` VARCHAR(255),
    `type` VARCHAR(50) DEFAULT 'TV' -- เพิ่มคอลัมน์ 'type' เพื่อระบุประเภทของช่อง (เช่น 'TV', 'Live Stream', 'Website')
);

-- 5. สร้างตาราง nationalities (สัญชาติ)
-- ตารางนี้เก็บข้อมูลสัญชาติสำหรับผู้เล่นและโค้ช
CREATE TABLE IF NOT EXISTS `nationalities` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `code` VARCHAR(10) UNIQUE,           -- รหัสสัญชาติ (เช่น TH, EN)
    `name` VARCHAR(255) NOT NULL UNIQUE, -- ชื่อสัญชาติ (เช่น ไทย, อังกฤษ)
    `flag_url` VARCHAR(255)              -- URL รูปภาพธงชาติ
);

-- 6. สร้างตาราง players (ผู้เล่น)
-- ตารางนี้เก็บข้อมูลและสถิติของผู้เล่น
CREATE TABLE IF NOT EXISTS `players` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `player_ref_id` INT UNIQUE,          -- ID อ้างอิงจาก API หรือแหล่งข้อมูลอื่น
    `league_id` INT,                     -- Foreign Key อ้างอิงไปยังตาราง leagues
    `team_id` INT,                       -- Foreign Key อ้างอิงไปยังตาราง teams
    `nationality_id` INT,                -- Foreign Key อ้างอิงไปยังตาราง nationalities
    `name` VARCHAR(255) NOT NULL,        -- ชื่อผู้เล่น
    `full_name_en` VARCHAR(255),         -- ชื่อเต็มภาษาอังกฤษ
    `shirt_number` INT,                  -- หมายเลขเสื้อ
    `position` VARCHAR(50),              -- ตำแหน่ง (เช่น FW, MF, DF, GK)
    `photo_url` VARCHAR(255),            -- URL รูปภาพผู้เล่น
    `matches_played` INT DEFAULT 0,      -- จำนวนนัดที่ลงเล่น
    `goals` INT DEFAULT 0,               -- จำนวนประตูที่ทำได้
    `yellow_cards` INT DEFAULT 0,        -- จำนวนใบเหลือง
    `red_cards` INT DEFAULT 0,           -- จำนวนใบแดง
    `status` INT DEFAULT 0,               -- สถานะผู้เล่น (เช่น active/inactive)
    FOREIGN KEY (`league_id`) REFERENCES `leagues`(`id`),
    FOREIGN KEY (`team_id`) REFERENCES `teams`(`id`),
    FOREIGN KEY (`nationality_id`) REFERENCES `nationalities`(`id`)
);

-- 7. สร้างตาราง coaches (โค้ช)
-- ตารางนี้เก็บข้อมูลโค้ช
CREATE TABLE IF NOT EXISTS `coaches` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `coach_ref_id` INT UNIQUE,           -- ID อ้างอิงจาก API หรือแหล่งข้อมูลอื่น
    `name` VARCHAR(255) NOT NULL,
    `birthday` DATE,                     -- วันเกิด
    `team_id` INT,                       -- Foreign Key อ้างอิงไปยังตาราง teams
    `nationality_id` INT,                -- Foreign Key อ้างอิงไปยังตาราง nationalities
    FOREIGN KEY (`team_id`) REFERENCES `teams`(`id`),
    FOREIGN KEY (`nationality_id`) REFERENCES `nationalities`(`id`)
);

-- 8. สร้างตาราง league_standings (ตารางคะแนนลีก)
-- ตารางนี้เก็บข้อมูลตารางคะแนน/อันดับของแต่ละทีมในแต่ละลีก
CREATE TABLE IF NOT EXISTS `standings` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `league_id` INT NOT NULL,           -- Foreign Key อ้างอิงไปยังตาราง leagues
    `team_id` INT NOT NULL,             -- Foreign Key อ้างอิงไปยังตาราง teams
    `stage_id` INT,                     -- เพิ่ม: FK ไปตาราง stage
    `matches_played` INT DEFAULT 0,     -- จำนวนนัดที่ลงแข่ง
    `wins` INT DEFAULT 0,               -- จำนวนนัดที่ชนะ
    `draws` INT DEFAULT 0,              -- จำนวนนัดที่เสมอ
    `losses` INT DEFAULT 0,             -- จำนวนนัดที่แพ้
    `goals_for` INT DEFAULT 0,          -- จำนวนประตูที่ยิงได้
    `goals_against` INT DEFAULT 0,      -- จำนวนประตูที่เสียไป
    `goal_difference` INT DEFAULT 0,    -- ผลต่างประตูได้เสีย
    `points` INT DEFAULT 0,             -- คะแนนรวม
    `current_rank` INT,                 -- อันดับปัจจุบัน
    `status` INT,                        -- รอบการแข่งขัน (ถ้ามี)
    UNIQUE (`league_id`, `team_id`, `stage_id`), -- กำหนดให้แต่ละทีมมีข้อมูลคะแนนเดียวในแต่ละลีกและสเตจ
    FOREIGN KEY (`league_id`) REFERENCES `leagues`(`id`),
    FOREIGN KEY (`team_id`) REFERENCES `teams`(`id`)
);



-- สร้างตาราง stage (โซน/ประเภทการแข่งขัน)
CREATE TABLE IF NOT EXISTS `stage` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `stage_name` VARCHAR(255) NOT NULL UNIQUE
);

-- 9. สร้างตาราง matches (แมตช์การแข่งขัน)
-- ตารางนี้เก็บข้อมูลการแข่งขันแต่ละนัด
CREATE TABLE IF NOT EXISTS `matches` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `match_ref_id` INT UNIQUE NOT NULL, -- ID อ้างอิงจาก API ที่ไม่ซ้ำกัน
    `start_date` DATE NOT NULL,
    `start_time` TIME NOT NULL,
    `league_id` INT,                    -- Foreign Key อ้างอิงไปยังตาราง leagues
    `stage_id` INT,                     -- FK ไปตาราง stage
    `stadium_id` INT,                   -- เพิ่ม: FK ไปตาราง stadiums
    `home_team_id` INT,                 -- Foreign Key อ้างอิงไปยังตาราง teams (ทีมเหย้า)
    `away_team_id` INT,                 -- Foreign Key อ้างอิงไปยังตาราง teams (ทีมเยือน)
    `channel_id` INT,                   -- Foreign Key อ้างอิงไปยังตาราง channels (ช่องทีวีหลัก)
    `live_channel_id` INT,              -- Foreign Key อ้างอิงไปยังตาราง channels (ช่องถ่ายทอดสด)
    `home_score` INT,                   -- คะแนนทีมเหย้า
    `away_score` INT,                   -- คะแนนทีมเยือน
    `match_status` VARCHAR(50),         -- สถานะการแข่งขัน (เช่น FIXTURE, FINISHED)
    FOREIGN KEY (`league_id`) REFERENCES `leagues`(`id`),
    FOREIGN KEY (`home_team_id`) REFERENCES `teams`(`id`),
    FOREIGN KEY (`away_team_id`) REFERENCES `teams`(`id`),
    FOREIGN KEY (`channel_id`) REFERENCES `channels`(`id`),
    FOREIGN KEY (`live_channel_id`) REFERENCES `channels`(`id`),
    FOREIGN KEY (`stage_id`) REFERENCES `stage`(`id`),
    FOREIGN KEY (`stadium_id`) REFERENCES `stadiums`(`id`)
);

-- 10. เพิ่ม Foreign Keys เพื่อแก้ไขปัญหา Circular Dependency

ALTER TABLE `stadiums` ADD CONSTRAINT `fk_stadiums_team_id` FOREIGN KEY (`team_id`) REFERENCES `teams`(`id`);
ALTER TABLE `teams` ADD CONSTRAINT `fk_teams_stadium_id` FOREIGN KEY (`stadium_id`) REFERENCES `stadiums`(`id`);

