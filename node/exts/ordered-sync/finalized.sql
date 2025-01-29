-- valid_low_points is a CTE that gets each pending_data row
-- where the previous point equals the last processed point.
-- This itself is not a recursive query, but it is used in a
-- recursive query in the next CTE.
{kwil_ordered_sync} WITH RECURSIVE valid_low_points AS (
    SELECT
        p.point,
        p.topic_id,
        p.previous_point,
        p.data,
        t.name
    FROM pending_data p
    JOIN topics t ON p.topic_id = t.id
     -- we use IS NOT DISTINCT FROM to handle NULLs, since the first data point will have a NULL previous_point and last_processed_point
    WHERE previous_point IS NOT DISTINCT FROM last_processed_point
), recursive_cte AS (
    -- base case: get the lowest point for each topic
    SELECT
        point,
        topic_id,
        previous_point,
        data,
        name
    FROM valid_low_points
    UNION ALL
    -- recursive case: get the next point for each topic
    SELECT
        pending_data.point,
        pending_data.topic_id,
        pending_data.previous_point,
        pending_data.data,
        recursive_cte.name
    FROM pending_data
    -- no need for IS DISTINCT FROM since only the first ever row can have a NULL previous_point,
    -- and we already have that row in the base case
    JOIN recursive_cte ON pending_data.previous_point = recursive_cte.point AND pending_data.topic_id = recursive_cte.topic_id
) SELECT point, previous_point, data, name FROM recursive_cte ORDER BY name, point;