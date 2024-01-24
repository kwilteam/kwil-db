package voting

/*
	sql.go contains information about the sql tables used by the voting system.
*/

// TODO: scope all table names with a scope like kwild_voting

const (
	votingSchemaName = `kwild_voting`

	createVotingSchema = `CREATE SCHEMA IF NOT EXISTS ` + votingSchemaName

	// tableResolutions is the sql table used to store resolutions that can be voted on
	tableResolutions = `CREATE TABLE IF NOT EXISTS ` + votingSchemaName + `.resolutions (
		id BYTEA PRIMARY KEY, -- id is an rfc4122 uuid derived from the body
		body BYTEA, -- body is the actual resolution info
		type BYTEA, -- type is the type of resolution
		voteBodyProposer BYTEA, -- voteBodyProposer is the identifier of the node that supplied the vote body
		expiration INT8 NOT NULL, -- expiration is the blockheight at which the resolution expires
		extraVoteID BOOLEAN NOT NULL DEFAULT FALSE, -- If voteBodyProposer had sent VoteID before VoteBody, this is set to true
		UNIQUE (id, body, type),
		FOREIGN KEY(type) REFERENCES ` + votingSchemaName + `.resolution_types(id) ON UPDATE CASCADE ON DELETE CASCADE,
		FOREIGN KEY(voteBodyProposer) REFERENCES ` + votingSchemaName + `.voters(id)
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

	// resolutionIDExists is the sql statement used to ensure a resolution ID is present in the resolutions table
	resolutionIDExists = `INSERT INTO ` + votingSchemaName + `.resolutions (id, expiration) VALUES ($1, $2)
		ON CONFLICT(id) DO NOTHING;`

	// upsertResolution is the sql statement used to ensure a resolution is
	// present in the resolutions table. In scenarios where VoteID is received
	// before VoteBody, the body and type will be updated in the existing
	// resolution entry.
	upsertResolution = `INSERT INTO ` + votingSchemaName + `.resolutions (id, body, type, expiration, voteBodyProposer, extraVoteID)
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
		voteBodyProposer = $5,
		extraVoteID = $6;`

	// upsertVoter is the sql statement used to ensure a voter is present in the voters table.  If the voter is present, the power is updated.
	upsertVoter = `INSERT INTO ` + votingSchemaName + `.voters (id, name, power) VALUES ($1, $2, $3)
		ON CONFLICT(id) DO UPDATE SET power = $3;`

	// removeVoter is the sql statement used to remove a voter from the voters table
	removeVoter = `DELETE FROM ` + votingSchemaName + `.voters WHERE id = $1;`

	// getVoterPower is the sql statement used to get the power and name of a voter
	getVoterPower = `SELECT power FROM ` + votingSchemaName + `.voters WHERE id = $1;`

	// getVoterName is the sql statement used to get the name of a voter
	getVoterName = `SELECT name FROM ` + votingSchemaName + `.voters WHERE id = $1;`

	// addVote adds a vote for a resolution
	addVote = `INSERT INTO ` + votingSchemaName + `.votes (resolution_id, voter_id) VALUES ($1, $2)
		ON CONFLICT(resolution_id, voter_id) DO NOTHING;`

	// hasVoted checks if a voter has voted on a resolution
	hasVoted = `SELECT resolution_id FROM ` + votingSchemaName + `.votes WHERE resolution_id = $1 AND voter_id = $2;`

	// expired Resolutions is the sql statement used to get expired resolutions
	expiredResolutions = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration, SUM(vr.power) AS approved_power, r.voteBodyProposer AS voteBodyProposer, r.extraVoteID AS extraVoteID
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.expiration <= $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.voteBodyProposer, r.extraVoteID;`

	// containsBody checks if a resolution has a body
	containsBody = `SELECT body is not null FROM ` + votingSchemaName + `.resolutions WHERE id = $1;`

	// getResolutionVoters is the sql statement used to get the voters info of a resolution
	getResolutionVoters = `SELECT vr.name, vr.power
	FROM ` + votingSchemaName + `.votes v
	JOIN ` + votingSchemaName + `.voters vr ON v.voter_id = vr.id
	WHERE v.resolution_id = $1;`

	// getResolutionVoteInfo is the sql statement used to get a resolution and the associated vote info.
	// while it would be nice to get the needed power as well, it is significantly more expensive to do so.
	// it would be better to cache the maximum needed power for a given resolution.
	getResolutionVoteInfo = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration, SUM(vr.power) AS approved_power, r.voteBodyProposer AS voteBodyProposer, r.extraVoteID AS extraVoteID
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.id = $1
	GROUP BY r.id, r.body, t.name, r.expiration, r.voteBodyProposer, r.extraVoteID;`

	// getUnfilledResolutionVoteInfo gets en expiration and approved power for a resolution that has not been filled with a body and type
	getUnfilledResolutionVoteInfo = `SELECT r.expiration AS expiration, SUM(vr.power) AS approved_power
	FROM ` + votingSchemaName + `.resolutions AS r
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.id = $1
	GROUP BY r.id, r.expiration;`

	// resolutionsByType is the sql statement used to get resolutions of a given type
	resolutionsByType = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	WHERE t.name = $1;`

	// getConfirmedResolutions is the statement used to get all resolutions
	// that have been confirmed above the given power threshold
	// we do not calculate the threshold here since we need to guarantee accuracy
	// using big ints.
	// it orders by id for determinism
	getConfirmedResolutions = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration, SUM(vr.power) AS approved_power, r.voteBodyProposer AS voteBodyProposer, r.extraVoteID AS extraVoteID
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.body IS NOT NULL
	GROUP BY r.id, r.body, t.name, r.expiration, r.voteBodyProposer, r.extraVoteID
	HAVING SUM(vr.power) >= $1
	ORDER BY r.id;`

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

	// markProcessed marks a resolution as processed
	markProcessed = `INSERT INTO ` + votingSchemaName + `.processed (id) VALUES ($1);`

	// alreadyProcessed checks if a resolution has already been processed
	alreadyProcessed = `SELECT id FROM ` + votingSchemaName + `.processed WHERE id = $1;`
)
