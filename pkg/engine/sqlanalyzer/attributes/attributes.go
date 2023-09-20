/*
Package attributes analyzes a returned relations attributes, maintaining order.
This is useful for determining the relation schema that a query / CTE returns.

For example, given the following query:

	WITH satoshi_posts AS (
		SELECT id, title, content FROM posts
		WHERE user_id = (
			SELECT id FROM users WHERE username = 'satoshi' LIMIT 1
		)
	)
	SELECT id, title FROM satoshi_posts;

The attributes package will be able to determine that:
 1. The result of this query is a relation with two attributes: id and title
 2. The result of the common table expression satoshi_posts is a relation with three attributes: id, title, and content
*/
package attributes
