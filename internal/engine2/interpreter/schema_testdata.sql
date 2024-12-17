--TODO: delete this file, it is just for testing

-- DELETE ME! just for testing
INSERT INTO kwild_engine.roles (name) VALUES ('admin') ON CONFLICT DO NOTHING;
INSERT INTO kwild_engine.role_inheritance (inheriter_id, inherited_from_id) VALUES (
    (SELECT id FROM kwild_engine.roles WHERE name = 'admin'),
    (SELECT id FROM kwild_engine.roles WHERE name = 'default')
) ON CONFLICT DO NOTHING;
INSERT INTO kwild_engine.privileges (privilege_type, role_id) VALUES ('INSERT', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'admin'
)), ('UPDATE', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'admin'
)) ON CONFLICT DO NOTHING;

-- leader role
INSERT INTO kwild_engine.roles (name) VALUES ('leader') ON CONFLICT DO NOTHING;
INSERT INTO kwild_engine.role_inheritance (inheriter_id, inherited_from_id) VALUES (
    (SELECT id FROM kwild_engine.roles WHERE name = 'leader'),
    (SELECT id FROM kwild_engine.roles WHERE name = 'admin')
) ON CONFLICT DO NOTHING;
INSERT INTO kwild_engine.privileges (privilege_type, role_id) VALUES ('DELETE', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'leader'
)), ('DROP', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'leader'
)) ON CONFLICT DO NOTHING;