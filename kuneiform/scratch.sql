
CREATE OR REPLACE FUNCTION create_post(_param_content text) 
RETURNS void AS $$
DECLARE
_param_post_count INT8;
_param_user_id uuid;
BEGIN

SELECT * INTO _param_user_id, _param_post_count from increment_post_count(current_setting('ctx.caller'));

raise notice 'demwldewnmwkl';

raise notice 'user id: %s',_param_user_id;

  INSERT  INTO "posts" ( "id" , "content" , "author_id" , "post_num" )  VALUES  (uuid_generate_v5  ('985b93a4-2045-44d6-bde4-442a4e498bc6' ::uuid  ,current_setting('ctx.txid')), _param_content, _param_user_id, _param_post_count); 

END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION increment_post_count(_param_address text, OUT _out_0 uuid, OUT _out_1 INT8) AS $$
DECLARE
_param_row RECORD;
BEGIN

FOR _param_row IN UPDATE  "users"  SET  "post_count" ="post_count"  +1  WHERE"users" . "address"  =_param_address  RETURNING  "users" . "id"  ,  "users" . "post_count"  LOOP
raise notice 'id from row: %s',_param_row.id;
_out_0 := _param_row.id;
_out_1 := _param_row.post_count;
RETURN;
END LOOP;

_out_0 := '985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid;
_out_1 := 1;

END;
$$ LANGUAGE plpgsql;