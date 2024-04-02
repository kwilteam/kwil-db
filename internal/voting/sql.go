package voting

/*
	sql.go contains information about the sql tables used by the voting system.
*/

/*
	Final schema after all the upgrades:
	resolutions:
		- id: uuid
		- body: bytea
		- type: bytea
		- vote_body_proposer: bytea
		- expiration: int8

	resolution_types:
		- id: uuid

	voters:
		- id: uuid
		- name: bytea
		- power: int8

	votes:
		- resolution_id: uuid
		- voter_id: uuid

	processed:
		- id: uuid
*/

const (
	votingSchemaName = `kwild_voting`

	voteStoreVersion = 2

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

	// upsertResolution is the sql statement used to ensure a resolution is
	// present in the resolutions table. In scenarios where VoteID is received
	// before VoteBody, the body, type and expiration will be updated in the existing
	// resolution entry.
	insertResolution = `INSERT INTO ` + votingSchemaName + `.resolutions (id, body, type, expiration, vote_body_proposer)
	VALUES ($1, $2, (
		SELECT id
		FROM ` + votingSchemaName + `.resolution_types
		WHERE name = $3
	), $4, $5) ON CONFLICT(id) DO NOTHING;` // should fail if the resolution already exists

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

	// deleteResolutions deletes a set of resolutions
	// it is meant to be used in formatResolutionList
	deleteResolutions = `DELETE FROM ` + votingSchemaName + `.resolutions WHERE id =ANY($1);` // $1 is a BYTEA[], unlike when using IN where you need a *list/set*

	// createResolutionType creates a resolution type
	createResolutionType = `INSERT INTO ` + votingSchemaName + `.resolution_types (id, name) VALUES ($1, $2)
		ON CONFLICT(id) DO NOTHING;`

	// markManyProcessed marks many resolutions as processed
	markManyProcessed = `INSERT INTO ` + votingSchemaName + `.processed (id) SELECT unnest($1::BYTEA[]);`

	// alreadyProcessed checks if a resolution has already been processed
	alreadyProcessed = `SELECT id FROM ` + votingSchemaName + `.processed WHERE id = $1;`

	// returnNotProcessed returns all resolutions in the input array that do not exist in the processed table
	returnProcessed = `SELECT id FROM ` + votingSchemaName + `.processed WHERE id =ANY($1);`

	// getResolutionsFullInfoByPower and getResolutionsFullInfoByExpiration are used to get the full info of a set of resolutions
	// they should be updated together if their return values change
	// gets the following info for a set of resolutions:
	// id, body, type, expiration, approved_power, voters (power concatted with name), vote_body_proposer
	// it is filtered by the approved power
	getResolutionsFullInfoByPower = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE t.name = $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer
	HAVING SUM(vr.power) >= $2
	ORDER BY r.id;` // order by id for determinism. ids are unique in the result.

	// GetResolutionsFullInfoByType gets the full info of a set of resolutions by type
	getResolutionsFullInfoByType = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE t.name = $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer
	ORDER BY r.id;` // order by id for determinism. ids are unique in the result.

	// gets the following info for a set of resolutions:
	// id, body, type, expiration, approved_power, voters (power concatted with name), vote_body_proposer
	// it is filtered by the expiration (less than or equal to)
	getResolutionsFullInfoByExpiration = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.expiration <= $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer
	ORDER BY r.id;` // order by id for determinism. ids are unique in the result.

	// getFullResolutionInfo gets the full info of a resolution.
	// it redundantly returns the id for convenience with functions that consume this query
	getFullResolutionInfo = `
	SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration,
		SUM(vr.power) AS approved_power, ARRAY_AGG(int8send(vr.power) || vr.name ORDER BY vr.id) AS voters,
		r.vote_body_proposer AS vote_body_proposer
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.id = $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.vote_body_proposer
	ORDER BY r.id;` // order by not necessary since only one result?

	allVoters                      = `SELECT name, power FROM ` + votingSchemaName + `.voters;`
	getResolutionByTypeAndProposer = `SELECT r.id FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	WHERE t.name = $1 AND vote_body_proposer = $2
	ORDER BY r.id;` // order by id for determinism
)

// upgrades V0 -> V1
const (
	dropHeightTable = `DROP TABLE IF EXISTS ` + votingSchemaName + `.height`
)

// upgrades V1 -> V2
const (
	dropExtraVoteID = `ALTER TABLE ` + votingSchemaName + `.resolutions DROP COLUMN extra_vote_id;`
)
