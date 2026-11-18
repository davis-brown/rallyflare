-- +goose Up
-- Suspending a player (handleAdminRemoveContent, entity_type 'user') writes an
-- audit row with action='suspend'. Widen the check constraint to allow it.
ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check
    CHECK (action IN ('promote', 'demote', 'clear_avatar', 'suspend'));

-- +goose Down
DELETE FROM admin_actions WHERE action = 'suspend';
ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check
    CHECK (action IN ('promote', 'demote', 'clear_avatar'));
