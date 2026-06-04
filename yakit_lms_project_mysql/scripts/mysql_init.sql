SET NAMES utf8mb4;
SET time_zone = '+09:00';

CREATE DATABASE IF NOT EXISTS yakit_lms
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE yakit_lms;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  full_name VARCHAR(160) NOT NULL,
  email VARCHAR(160) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(24) NOT NULL DEFAULT 'student',
  group_name VARCHAR(80) NULL,
  specialty VARCHAR(220) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY idx_users_email (email),
  KEY idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS courses (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(180) NOT NULL,
  category VARCHAR(80) NULL,
  description TEXT NULL,
  image_url VARCHAR(500) NULL,
  teacher_id BIGINT UNSIGNED NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_courses_teacher_id (teacher_id),
  CONSTRAINT fk_courses_teacher FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS enrollments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  course_id BIGINT UNSIGNED NOT NULL,
  progress INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY idx_enrollment (user_id, course_id),
  KEY idx_enrollments_course_id (course_id),
  CONSTRAINT fk_enrollments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_enrollments_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS course_modules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  course_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(180) NOT NULL,
  description TEXT NULL,
  sort_order INT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_course_modules_course_id (course_id),
  CONSTRAINT fk_course_modules_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS materials (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  course_id BIGINT UNSIGNED NOT NULL,
  module_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(180) NOT NULL,
  kind VARCHAR(40) NOT NULL DEFAULT 'link',
  url VARCHAR(500) NULL,
  content TEXT NULL,
  sort_order INT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_materials_course_id (course_id),
  KEY idx_materials_module_id (module_id),
  CONSTRAINT fk_materials_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_materials_module FOREIGN KEY (module_id) REFERENCES course_modules(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS assignments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  course_id BIGINT UNSIGNED NOT NULL,
  module_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(180) NOT NULL,
  description TEXT NULL,
  instruction_url VARCHAR(500) NULL,
  kind VARCHAR(40) NOT NULL DEFAULT 'assignment',
  deadline DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  max_score INT NOT NULL DEFAULT 5,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_assignments_course_id (course_id),
  KEY idx_assignments_module_id (module_id),
  CONSTRAINT fk_assignments_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_assignments_module FOREIGN KEY (module_id) REFERENCES course_modules(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS submissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  assignment_id BIGINT UNSIGNED NOT NULL,
  student_id BIGINT UNSIGNED NOT NULL,
  answer TEXT NULL,
  file_url VARCHAR(500) NULL,
  file_name VARCHAR(260) NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'pending',
  grade INT NULL,
  comment TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_submissions_assignment_id (assignment_id),
  KEY idx_submissions_student_id (student_id),
  CONSTRAINT fk_submissions_assignment FOREIGN KEY (assignment_id) REFERENCES assignments(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_submissions_student FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS submission_files (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  submission_id BIGINT UNSIGNED NOT NULL,
  file_url VARCHAR(500) NOT NULL,
  file_name VARCHAR(260) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_submission_files_submission_id (submission_id),
  CONSTRAINT fk_submission_files_submission FOREIGN KEY (submission_id) REFERENCES submissions(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tests (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  course_id BIGINT UNSIGNED NOT NULL,
  module_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(180) NOT NULL,
  description TEXT NULL,
  deadline DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  max_score INT NOT NULL DEFAULT 5,
  attempt_limit INT NOT NULL DEFAULT 1,
  duration_minutes INT NOT NULL DEFAULT 0,
  show_result TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_tests_course_id (course_id),
  KEY idx_tests_module_id (module_id),
  CONSTRAINT fk_tests_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_tests_module FOREIGN KEY (module_id) REFERENCES course_modules(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS test_questions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  test_id BIGINT UNSIGNED NOT NULL,
  question TEXT NOT NULL,
  type VARCHAR(40) NOT NULL DEFAULT 'radio',
  image_url VARCHAR(500) NULL,
  options_json TEXT NULL,
  correct_answer VARCHAR(500) NULL,
  correct_answers_json TEXT NULL,
  points INT NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_test_questions_test_id (test_id),
  CONSTRAINT fk_test_questions_test FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS test_attempts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  test_id BIGINT UNSIGNED NOT NULL,
  student_id BIGINT UNSIGNED NOT NULL,
  answers TEXT NULL,
  attempt_number INT NOT NULL DEFAULT 1,
  score INT NOT NULL DEFAULT 0,
  auto_score INT NOT NULL DEFAULT 0,
  max_score INT NOT NULL DEFAULT 0,
  status VARCHAR(40) NOT NULL DEFAULT 'completed',
  comment TEXT NULL,
  manual_grades TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_test_attempts_test_id (test_id),
  KEY idx_test_attempts_student_id (student_id),
  CONSTRAINT fk_test_attempts_test FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_test_attempts_student FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS grades (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  course_id BIGINT UNSIGNED NOT NULL,
  kind VARCHAR(40) NOT NULL,
  assignment_id BIGINT UNSIGNED NULL,
  test_id BIGINT UNSIGNED NULL,
  test_attempt_id BIGINT UNSIGNED NULL,
  value INT NOT NULL,
  max_value INT NOT NULL DEFAULT 5,
  comment TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_grades_user_id (user_id),
  KEY idx_grades_course_id (course_id),
  KEY idx_grades_kind (kind),
  KEY idx_grades_assignment_id (assignment_id),
  KEY idx_grades_test_id (test_id),
  KEY idx_grades_test_attempt_id (test_attempt_id),
  CONSTRAINT fk_grades_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_grades_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_grades_assignment FOREIGN KEY (assignment_id) REFERENCES assignments(id) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT fk_grades_test FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT fk_grades_test_attempt FOREIGN KEY (test_attempt_id) REFERENCES test_attempts(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS news (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(220) NOT NULL,
  excerpt VARCHAR(400) NULL,
  content TEXT NULL,
  image_url VARCHAR(500) NULL,
  published_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS teacher_profiles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(160) NOT NULL,
  position VARCHAR(180) NULL,
  email VARCHAR(160) NULL,
  image_url VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS applications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  first_name VARCHAR(80) NOT NULL,
  last_name VARCHAR(80) NOT NULL,
  phone VARCHAR(40) NOT NULL,
  direction VARCHAR(260) NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'new',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS schedule_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  group_name VARCHAR(80) NULL,
  course_id BIGINT UNSIGNED NULL,
  course_title VARCHAR(180) NOT NULL,
  teacher_name VARCHAR(160) NULL,
  room VARCHAR(80) NULL,
  starts_at DATETIME(3) NOT NULL,
  ends_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_schedule_items_group_name (group_name),
  KEY idx_schedule_items_course_id (course_id),
  KEY idx_schedule_items_starts_at (starts_at),
  CONSTRAINT fk_schedule_items_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Демо-пароль для всех учетных записей: 123456
SET @demo_password_hash = '$2a$10$NtIEkUnFGX75PQm.TagyPOgU/mOAEfZrSGSAJiz9mfHnICwvhqbf.';

INSERT IGNORE INTO users (id, full_name, email, password_hash, role, group_name, specialty) VALUES
(1, 'Александр Титович', 'student@yakit.ru', @demo_password_hash, 'student', 'КИСП-23', '09.02.07 Информационные системы и программирование'),
(2, 'Никита Рожин', 'student2@yakit.ru', @demo_password_hash, 'student', 'КИСП-23', '09.02.07 Информационные системы и программирование'),
(3, 'Дьулуур Федоров', 'teacher@yakit.ru', @demo_password_hash, 'teacher', NULL, 'Преподаватель отделения ИТ'),
(4, 'Елена Белова', 'teacher2@yakit.ru', @demo_password_hash, 'teacher', NULL, 'Преподаватель UI/UX'),
(5, 'Иван Пронин', 'admin@yakit.ru', @demo_password_hash, 'admin', NULL, 'Заведующий отделением ИТ');

INSERT IGNORE INTO teacher_profiles (id, name, position, email, image_url) VALUES
(1, 'Дьулуур Федоров', 'Преподаватель веб-разработки', 'teacher@yakit.ru', 'https://images.unsplash.com/photo-1560250097-0b93528c311a?auto=format&fit=crop&q=80&w=600'),
(2, 'Елена Белова', 'Преподаватель UI/UX', 'teacher2@yakit.ru', 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=600'),
(3, 'Игорь Марков', 'Backend архитектор', 'markov@yakit.ru', 'https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?auto=format&fit=crop&q=80&w=600');

INSERT IGNORE INTO news (id, title, excerpt, content, image_url, published_at) VALUES
(1, 'Открытие новой IT-лаборатории', 'В колледже открылась лаборатория для проектной работы и практики по веб-разработке.', 'Новая лаборатория позволит студентам выполнять командные проекты, работать с серверным оборудованием и развивать навыки промышленной разработки.', 'https://images.unsplash.com/photo-1497366754035-f200968a6e72?auto=format&fit=crop&q=80&w=900', NOW() - INTERVAL 2 DAY),
(2, 'Студенты ЯКИТ защитили учебные проекты', 'Команды представили веб-приложения, backend-сервисы и интерфейсы для учебных задач.', 'На защите были представлены проекты на Vue.js, Go, MySQL и других современных технологиях.', 'https://images.unsplash.com/photo-1552664730-d307ca884978?auto=format&fit=crop&q=80&w=900', NOW() - INTERVAL 5 DAY),
(3, 'День открытых дверей', 'Абитуриенты смогут познакомиться с направлениями подготовки и подать заявление онлайн.', 'В программе: экскурсии по аудиториям, консультации преподавателей и демонстрация студенческих проектов.', 'https://images.unsplash.com/photo-1523240795612-9a054b0db644?auto=format&fit=crop&q=80&w=900', NOW() - INTERVAL 10 DAY);

INSERT IGNORE INTO courses (id, title, category, description, image_url, teacher_id, status) VALUES
(1, 'Веб-разработка на Vue.js', 'Frontend', 'Разработка интерфейсов, компонентный подход, реактивность и работа с REST API.', 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?auto=format&fit=crop&q=80&w=900', 3, 'active'),
(2, 'Backend на Go', 'Backend', 'Создание REST API, авторизация, работа с базой данных и архитектура серверного приложения.', 'https://images.unsplash.com/photo-1558494949-ef010cbdcc48?auto=format&fit=crop&q=80&w=900', 3, 'active'),
(3, 'UI/UX проектирование', 'Design', 'Основы интерфейсов, пользовательские сценарии и прототипирование.', 'https://images.unsplash.com/photo-1545235617-9465d2a55698?auto=format&fit=crop&q=80&w=900', 4, 'active');

INSERT IGNORE INTO course_modules (id, course_id, title, description, sort_order) VALUES
(1, 1, 'Компонентная архитектура', 'Основы Vue, структура компонента, состояние и события.', 1),
(2, 1, 'Интеграция с backend', 'REST API, JWT и обработка ошибок.', 2),
(3, 2, 'REST API на Go', 'Gin, middleware, маршрутизация и структура проекта.', 1),
(4, 2, 'База данных', 'GORM, MySQL, миграции и связи.', 2),
(5, 3, 'Проектирование интерфейса', 'Карточки, сетки, адаптивность, визуальная иерархия.', 1);

INSERT IGNORE INTO materials (id, course_id, module_id, title, kind, url, content, sort_order) VALUES
(1, 1, 1, 'Конспект: компоненты Vue', 'text', NULL, 'Компонент — независимый блок интерфейса, имеющий собственное состояние и шаблон.', 1),
(2, 1, 2, 'Документация по fetch', 'link', 'https://developer.mozilla.org/ru/docs/Web/API/Fetch_API', NULL, 1),
(3, 2, 3, 'Пример структуры Gin проекта', 'text', NULL, 'Маршруты, обработчики и middleware должны быть разделены по ответственности.', 1);

INSERT IGNORE INTO assignments (id, course_id, module_id, title, description, instruction_url, kind, deadline, max_score) VALUES
(1, 1, 1, 'Создать компонент карточки курса', 'Сделать Vue-компонент карточки курса с props: title, teacherName, progress.', '', 'assignment', NOW() + INTERVAL 7 DAY, 5),
(2, 1, 2, 'Подключить frontend к API', 'Получить список курсов через REST API и вывести их на странице.', '', 'assignment', NOW() + INTERVAL 10 DAY, 5),
(3, 2, 3, 'Создать CRUD для курсов', 'Реализовать endpoint создания курса и получение списка курсов.', '', 'assignment', NOW() + INTERVAL 14 DAY, 5);

INSERT IGNORE INTO tests (id, course_id, module_id, title, description, deadline, max_score, attempt_limit, duration_minutes, show_result) VALUES
(1, 1, 1, 'Тест по основам Vue', 'Проверка базовых понятий Vue.js.', NOW() + INTERVAL 9 DAY, 3, 2, 20, 1),
(2, 2, 3, 'Тест по REST API', 'Проверка понимания HTTP-методов и JSON.', NOW() + INTERVAL 12 DAY, 3, 2, 20, 1);

INSERT IGNORE INTO test_questions (id, test_id, question, type, image_url, options_json, correct_answer, correct_answers_json, points, sort_order) VALUES
(1, 1, 'Как называется основная сущность интерфейса во Vue?', 'radio', '', '["Компонент","Контроллер","Миграция"]', 'Компонент', '["Компонент"]', 1, 1),
(2, 1, 'Какие форматы подходят для обмена frontend и backend?', 'checkbox', '', '["JSON","XML","DOCX"]', 'JSON', '["JSON","XML"]', 1, 2),
(3, 2, 'Какой HTTP-метод обычно используется для создания записи?', 'radio', '', '["POST","GET","DELETE"]', 'POST', '["POST"]', 1, 1),
(4, 2, 'Что хранит JWT?', 'text', '', '[]', 'Подписанные claims', '["Подписанные claims"]', 1, 2);

INSERT IGNORE INTO enrollments (id, user_id, course_id, progress) VALUES
(1, 1, 1, 55),
(2, 1, 2, 25),
(3, 1, 3, 15),
(4, 2, 1, 45),
(5, 2, 2, 30);

INSERT IGNORE INTO submissions (id, assignment_id, student_id, answer, status, grade, comment) VALUES
(1, 1, 1, 'Компонент карточки курса создан, код приложен ссылкой.', 'graded', 5, 'Хорошая работа');

INSERT IGNORE INTO grades (id, user_id, course_id, kind, assignment_id, test_id, test_attempt_id, value, max_value, comment) VALUES
(1, 1, 1, 'assignment', 1, NULL, NULL, 5, 5, 'Хорошая работа');

INSERT IGNORE INTO schedule_items (id, group_name, course_id, course_title, teacher_name, room, starts_at, ends_at) VALUES
(1, 'КИСП-23', 1, 'Веб-разработка на Vue.js', 'Дьулуур Федоров', 'Ауд. 305', TIMESTAMP(CURDATE(), '09:00:00'), TIMESTAMP(CURDATE(), '10:30:00')),
(2, 'КИСП-23', 2, 'Backend на Go', 'Дьулуур Федоров', 'Лаб. 2', TIMESTAMP(CURDATE(), '10:40:00'), TIMESTAMP(CURDATE(), '12:10:00')),
(3, 'КИСП-23', 3, 'UI/UX проектирование', 'Елена Белова', 'Ауд. 401', TIMESTAMP(CURDATE(), '12:40:00'), TIMESTAMP(CURDATE(), '14:10:00'));
