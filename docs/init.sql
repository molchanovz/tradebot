-- Добавляем кабинеты (торговые площадки)
INSERT INTO "cabinets" ("name", "clientId", "key", "marketplace", "type", "sheetLink") VALUES
                                                                                           ('Wildberries FBO', 'WB123456', 'secure_key_1', 'WB', 'fbo', 'https://docs.google.com/spreadsheets/d/wb_fbo_1'),
                                                                                           ('Wildberries FBS', 'WB654321', 'secure_key_2', 'WB', 'fbs', 'https://docs.google.com/spreadsheets/d/wb_fbs_1'),
                                                                                           ('Ozon Основной', 'OZ789012', 'secure_key_3', 'OZON', 'all', 'https://docs.google.com/spreadsheets/d/oz_1'),
                                                                                           ('Яндекс.Маркет', 'YM345678', 'secure_key_4', 'YANDEX', 'fbs', 'https://docs.google.com/spreadsheets/d/ym_1');

-- Добавляем пользователей
INSERT INTO "users" ("tgId", "statusId", "isAdmin", "cabinetIds") VALUES
                                                                      (123456789, 1, true, ARRAY[1, 2, 3, 4]),  -- Администратор с доступом ко всем кабинетам
                                                                      (987654321, 1, false, ARRAY[1, 3]),       -- Обычный пользователь с доступом к WB FBO и Ozon
                                                                      (555555555, 1, false, ARRAY[2]);           -- Пользователь только с доступом к WB FBS

-- Добавляем данные о стоках
INSERT INTO "stocks" ("article", "updatedAt", "countFbo", "countFbs", "cabinetId") VALUES
                                                                                       ('ABC123', NOW() - INTERVAL '1 day', 150, 45, 1),
                                                                                       ('ABC123', NOW() - INTERVAL '2 hours', NULL, 30, 2),
                                                                                       ('XYZ789', NOW() - INTERVAL '3 hours', 80, NULL, 1),
                                                                                       ('DEF456', NOW() - INTERVAL '5 hours', 200, 60, 3),
                                                                                       ('GHI789', NOW(), NULL, 15, 4),
                                                                                       ('JKL012', NOW() - INTERVAL '1 hour', 90, 25, 1);

-- Добавляем заказы
INSERT INTO "orders" ("postingNumber", "article", "count", "cabinetId", "createdAt") VALUES
                                                                                         ('WB-123456', 'ABC123', 5, 1, NOW() - INTERVAL '3 days'),
                                                                                         ('WB-654321', 'XYZ789', 2, 1, NOW() - INTERVAL '2 days'),
                                                                                         ('OZ-789012', 'DEF456', 10, 3, NOW() - INTERVAL '1 day'),
                                                                                         ('YM-345678', 'GHI789', 3, 4, NOW() - INTERVAL '12 hours'),
                                                                                         ('WB-987654', 'ABC123', 7, 2, NOW() - INTERVAL '6 hours'),
                                                                                         ('WB-555555', 'JKL012', 4, 1, NOW() - INTERVAL '2 hours');
