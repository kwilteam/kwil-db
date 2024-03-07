package voting

/*
	sql.go contains information about the sql tables used by the voting system.
*/

const (
	votingSchemaName = `kwild_voting`

	voteStoreVersion = 0

	// tableResolutions is the sql table used to store resolutions that can be voted on.
	// the vote_body_proposer is the BYTEA of the public key of the submitter, NOT the UUID
	tableResolutions = `CREATE TABLE IF NOT EXISTS ` + votingSchemaName + `.resolutions (
		id BYTEA PRIMARY KEY, -- id is an rfc4122 uuid derived from the body
		body BYTEA, -- body is the actual resolution info
		type BYTEA, -- type is the type of resolution
		vote_body_proposer BYTEA, -- vote_body_proposer is the identifier of the node that supplied the vote body
		expiration INT8 NOT NULL, -- expiration is the blockheight at which the resolution expires
		extra_vote_id BOOLEAN NOT NULL DEFAULT FALSE, -- If vote_body_proposer had sent VoteID before VoteBody, this is set to true
		UNIQUE (id, body, type),
		FOREIGN KEY(type) REFERENCES ` + votingSchemaName + `.resolution_types(id) ON UPDATE CASCADE ON DELETE CASCADE
	);`

	// resolutionTypeIndex is the sql index used to index the type of a resolution
	resolutionsTypeIndex = `CREATE INDEX IF NOT EXISTS type_index ON ` + votingSchemaName + `.resolutions (type);`
	// resolution_types.type is already indexed...

	tableResolutionTypes = `CREATE TABLE IF NOT EXISTS ` + votingSchemaName + `.resolution_types (
		id BYTEA PRIMARY KEY, -- id is an rfc4122 uuid derived from the name
		name TEXT UNIQUE NOT NULL -- name is the name of the resolution type
	);`

	tableVoters = `CREATE TABLE IF NOT EXISTS ` + votingSchemaName + `.voters (
		id BYTEA PRIMARY KEY, -- id is an rfc4122 uuid derived from the voter
		name BYTEA UNIQUE NOT NULL, -- voter is the identifier of the voter
		power INT8 NOT NULL CHECK(power > 0) -- power is the voting power of the voter
	);`

	// votes tracks whether a voter has voted on a resolution
	tableVotes = `CREATE TABLE IF NOT EXISTS ` + votingSchemaName + `.votes (
		resolution_id BYTEA NOT NULL, 
		voter_id BYTEA NOT NULL,
		FOREIGN KEY(resolution_id) REFERENCES ` + votingSchemaName + `.resolutions(id) ON UPDATE CASCADE ON DELETE CASCADE,
		FOREIGN KEY(voter_id) REFERENCES ` + votingSchemaName + `.voters(id) ON UPDATE CASCADE ON DELETE CASCADE,
		PRIMARY KEY(resolution_id, voter_id) -- makes compound unique index
	);`

	// tableProcessed contains all processed resolution ids
	tableProcessed = `CREATE TABLE IF NOT EXISTS ` + votingSchemaName + `.processed (
		id BYTEA PRIMARY KEY
	);`

	tableHeight = `CREATE TABLE IF NOT EXISTS ` + votingSchemaName + `.height (
		name TEXT PRIMARY KEY, -- name is 'height'
		height INT NOT NULL
	);`

	getHeight = `SELECT height FROM ` + votingSchemaName + `.height WHERE name = 'height';`

	updateHeight = `INSERT INTO ` + votingSchemaName + `.height  (name, height) VALUES ('height', $1) 
		ON CONFLICT (name) DO UPDATE SET height = $1;`

	// ensureResolutionIDExists is the sql statement used to ensure a resolution ID is present in the resolutions table
	ensureResolutionIDExists = `INSERT INTO ` + votingSchemaName + `.resolutions (id, expiration) VALUES ($1, $2)
		ON CONFLICT(id) DO NOTHING;`

	// upsertResolution is the sql statement used to ensure a resolution is
	// present in the resolutions table. In scenarios where VoteID is received
	// before VoteBody, the body, type and expiration will be updated in the existing
	// resolution entry.
	upsertResolution = `INSERT INTO ` + votingSchemaName + `.resolutions (id, body, type, expiration, vote_body_proposer, extra_vote_id)
	VALUES ($1, $2, (
		SELECT id
		FROM ` + votingSchemaName + `.resolution_types
		WHERE name = $3
	), $4, $5, $6)
	ON CONFLICT(id)
		DO UPDATE
		SET body = $2, type = (
			SELECT id
			FROM ` + votingSchemaName + `.resolution_types
			WHERE name = $3
		),
		expiration = $4,
		vote_body_proposer = $5,
		extra_vote_id = $6;`

	// upsertVoter is the sql statement used to ensure a voter is present in the voters table.  If the voter is present, the power is updated.
	upsertVoter = `INSERT INTO ` + votingSchemaName + `.voters (id, name, power) VALUES ($1, $2, $3)
		ON CONFLICT(id) DO UPDATE SET power = $3;`

	// removeVoter is the sql statement used to remove a voter from the voters table
	removeVoter = `DELETE FROM ` + votingSchemaName + `.voters WHERE id = $1;`

	// getVoterPower is the sql statement used to get the power and name of a voter
	getVoterPower = `SELECT power FROM ` + votingSchemaName + `.voters WHERE id = $1;`

	// addVote adds a vote for a resolution
	addVote = `INSERT INTO ` + votingSchemaName + `.votes (resolution_id, voter_id) VALUES ($1, $2)
		ON CONFLICT(resolution_id, voter_id) DO NOTHING;`

	// hasVoted checks if a voter has voted on a resolution
	hasVoted = `SELECT resolution_id FROM ` + votingSchemaName + `.votes WHERE resolution_id = $1 AND voter_id = $2;`

	// containsBody checks if a resolution has a body
	containsBody = `SELECT body is not null FROM ` + votingSchemaName + `.resolutions WHERE id = $1;`
	// containsBodyArray checks if a set of resolutions have bodies.
	// it will return as many rows as the input array has elements, with a boolean for each.
	containsBodyArray = `WITH input AS (SELECT unnest($1::BYTEA[]) AS id)
	SELECT i.id, r.body IS NOT NULL AS contains_body
	FROM input AS i
	LEFT JOIN ` + votingSchemaName + `.resolutions AS r ON i.id = r.id;`

	// deleteResolutions deletes a set of resolutions
	// it is meant to be used in formatResolutionList
	deleteResolutions = `DELETE FROM ` + votingSchemaName + `.resolutions WHERE id =ANY($1);` // $1 is a BYTEA[], unlike when using IN where you need a *list/set*

	// totalPower gets the total power of all voters
	totalPower = `SELECT SUM(power) AS required_power FROM ` + votingSchemaName + `.voters;` // note: sum ( bigint ) → numeric
	// https://www.postgresql.org/docs/current/functions-aggregate.html#FUNCTIONS-AGGREGATE
	// The returned value will be a `numeric` and will scan as a pgtypes.Numeric (with pgx)

	// createResolutionType creates a resolution type
	createResolutionType = `INSERT INTO ` + votingSchemaName + `.resolution_types (id, name) VALUES ($1, $2)
		ON CONFLICT(id) DO NOTHING;`

	// markManyProcessed marks many resolutions as processed
	markManyProcessed = `INSERT INTO ` + votingSchemaName + `.processed (id) SELECT unnest($1::BYTEA[]);`

	// alreadyProcessed checks if a resolution has already been processed
	alreadyProcessed = `SELECT id FROM ` + votingSchemaName + `.processed WHERE id = $1;`

	// manyProcessed checks if many resolutions have already been processed
	// it will return a boolean for each resolution in the input array
	manyProcessed = `WITH input AS (SELECT unnest($1::BYTEA[]) AS id)
	SELECT i.id, p.id IS NOT NULL AS processed
	FROM input AS i
	LEFT JOIN ` + votingSchemaName + `.processed AS p ON i.id = p.id;`

	// getResolutionsFullInfoByPower and getResolutionsFullInfoByExpiration are used to get the full info of a set of resolutions
	// they should be updated together if their return values change

	// gets the following info for a set of resolutions:
	// id, body, type, expiration, approved_power, voters (power concatted with name), vote_body_proposer, extra_vote_id
	// it is filtered by the approved power
	getResolutionsFullInfoByPower = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer, r.extra_vote_id AS extra_vote_id
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE t.name = $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer, r.extra_vote_id
	HAVING SUM(vr.power) >= $2
	ORDER BY r.id;` // order by id for determinism. ids are unique in the result.

	// GetResolutionsFullInfoByType gets the full info of a set of resolutions by type
	getResolutionsFullInfoByType = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer, r.extra_vote_id AS extra_vote_id
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE t.name = $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer, r.extra_vote_id
	ORDER BY r.id;` // order by id for determinism. ids are unique in the result.

	// gets the following info for a set of resolutions:
	// id, body, type, expiration, approved_power, voters (power concatted with name), vote_body_proposer, extra_vote_id
	// it is filtered by the expiration (less than or equal to)
	getResolutionsFullInfoByExpiration = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer, r.extra_vote_id AS extra_vote_id
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.expiration <= $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer, r.extra_vote_id
	ORDER BY r.id;` // order by id for determinism. ids are unique in the result.

	// getFullResolutionInfo gets the full info of a resolution.
	// it redundantly returns the id for convenience with functions that consume this query
	getFullResolutionInfo = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer, r.extra_vote_id AS extra_vote_id
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.id = $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer, r.extra_vote_id
	ORDER BY r.id;` // order by not necessary since only one result?

	allVoters                      = `SELECT name, power FROM ` + votingSchemaName + `.voters;`
	getResolutionByTypeAndProposer = `SELECT r.id FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	WHERE t.name = $1 AND vote_body_proposer = $2
	ORDER BY r.id;` // order by id for determinism
)
