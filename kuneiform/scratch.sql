CREATE OR REPLACE FUNCTION ds_x975a5d686cb685d60bed9d99a0d3cb80e4f712dbf8d2700297085752.get_recent_posts_by_size(_param_username TEXT, _param_size INT8, _param_limit INT8) 
RETURNS TABLE(id UUID, content TEXT) AS $$
DECLARE
_param_count INT8;
_param_row RECORD;
BEGIN
IF _param_limit > 50 THEN
_param_limit := 50;
END IF;
_param_count := 0;
FOR _param_row IN SELECT * FROM ds_x975a5d686cb685d60bed9d99a0d3cb80e4f712dbf8d2700297085752.get_recent_posts(_param_username) LOOP
IF _param_count = _param_limit THEN
EXIT;
END IF;
IF length(_param_row.content) >= _param_size THEN
_param_count := _param_count + 1;
id := _param_row.id;
content := _param_row.content;
RETURN NEXT;
END IF;
END LOOP;

END;
$$ LANGUAGE plpgsql;