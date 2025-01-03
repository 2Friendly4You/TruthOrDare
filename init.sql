CREATE DATABASE IF NOT EXISTS truth_or_dare_db
CHARACTER SET = utf8mb4
COLLATE = utf8mb4_unicode_ci;

USE truth_or_dare_db;

CREATE TABLE IF NOT EXISTS questions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    language VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    type ENUM('truth', 'dare') NOT NULL,
    task TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL
);

CREATE TABLE IF NOT EXISTS tags (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS question_tags (
    question_id INT NOT NULL,
    tag_id INT NOT NULL,
    FOREIGN KEY (question_id) REFERENCES questions(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id),
    PRIMARY KEY (question_id, tag_id)
);

INSERT INTO questions (language, type, task) VALUES
    ('en', 'truth', 'Have you ever lied to your best friend?'),
    ('en', 'dare', 'Take a shot of vodka.'),
    ('de', 'truth', 'Hast du jemals Essen gestohlen?'),
    ('de', 'dare', 'Tanze für 1 Minute ohne Musik.');

INSERT INTO tags (name) VALUES
    ('18+'), ('alcohol'), ('food');

INSERT INTO question_tags (question_id, tag_id) VALUES
    (1, 1), (2, 2), (2, 1), (3, 3), (4, 1);
