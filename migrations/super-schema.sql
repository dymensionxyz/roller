-- trigger function for removing partition_table_info record after dropping the partitioned table
CREATE OR REPLACE FUNCTION func_trigger_00100_after_drop_table() RETURNS EVENT_TRIGGER AS $$
DECLARE
_dropped record;
BEGIN
FOR _dropped IN
SELECT schema_name, object_name
FROM pg_catalog.pg_event_trigger_dropped_objects()
WHERE object_type = 'table'
LOOP
IF _dropped.schema_name = 'public' THEN
EXECUTE 'DELETE FROM partition_table_info WHERE partition_table_name = $1' USING _dropped.object_name;
END IF;
END LOOP;
END;$$ LANGUAGE plpgsql;
CREATE EVENT TRIGGER trigger_00100_after_drop_table ON sql_drop
EXECUTE FUNCTION func_trigger_00100_after_drop_table();
