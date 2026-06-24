
ALTER TABLE reserved_handles ENABLE ROW LEVEL SECURITY;
CREATE POLICY reserved_handles_select_all ON reserved_handles FOR SELECT USING (true);
