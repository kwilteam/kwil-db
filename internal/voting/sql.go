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
		expiration INT8 NOT NULL, -- expiration is the blockheight at which the resolution expires
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

	// resolutionIDExists is the sql statement used to ensure a resolution ID is present in the resolutions table
	resolutionIDExists = `INSERT INTO ` + votingSchemaName + `.resolutions (id, expiration) VALUES ($1, $2)
		ON CONFLICT(id) DO NOTHING;`

	// upsertResolution is the sql statement used to ensure a resolution is
	// present in the resolutions table. In scenarios where VoteID is received
	// before VoteBody, the body and type will be updated in the existing
	// resolution entry.
	upsertResolution = `INSERT INTO ` + votingSchemaName + `.resolutions (id, body, type, expiration)
	VALUES ($1, $2, (
		SELECT id
		FROM ` + votingSchemaName + `.resolution_types
		WHERE name = $3
	), $4)
	ON CONFLICT(id)
		DO UPDATE
		SET body = $2, type = (
			SELECT id
			FROM ` + votingSchemaName + `.resolution_types
			WHERE name = $3
		);`

	// upsertVoter is the sql statement used to ensure a voter is present in the voters table.  If the voter is present, the power is updated.
	upsertVoter = `INSERT INTO ` + votingSchemaName + `.voters (id, name, power) VALUES ($1, $2, $3)
		ON CONFLICT(id) DO UPDATE SET power = EXCLUDED.power + $3;` // IS THIS INCREMENT REALLY RIGHT OR SHOULD THIS BE power = $3 ?

	// decreaseVoterPower is the sql statement used to decrease the power of a voter
	// this is necessary because the voters table CHECK filters before the on conflict
	decreaseVoterPower = `UPDATE ` + votingSchemaName + `.voters SET power = power - $2 WHERE id = $1;`

	// removeVoter is the sql statement used to remove a voter from the voters table
	removeVoter = `DELETE FROM ` + votingSchemaName + `.voters WHERE id = $1;`

	// getVoterPower is the sql statement used to get the power and name of a voter
	getVoterPower = `SELECT power FROM ` + votingSchemaName + `.voters WHERE id = $1;`

	// addVote adds a vote for a resolution
	addVote = `INSERT INTO ` + votingSchemaName + `.votes (resolution_id, voter_id) VALUES ($1, $2)
		ON CONFLICT(resolution_id, voter_id) DO NOTHING;`

	// hasVoted checks if a voter has voted on a resolution
	hasVoted = `SELECT resolution_id FROM ` + votingSchemaName + `.votes WHERE resolution_id = $1 AND voter_id = $2;`

	// expireResolutions is the sql statement used to expire resolutions
	// it will expire resolutions that have an expiration less than or equal to the given blockheight
	expireResolutions = `DELETE FROM ` + votingSchemaName + `.resolutions WHERE expiration <= $1 RETURNING id;`

	// getResolutionBody gets a resolution body by id
	getResolutionBody = `SELECT body FROM ` + votingSchemaName + `.resolutions WHERE id = $1;`

	// getResolutionVoteInfo is the sql statement used to get a resolution and the associated vote info.
	// while it would be nice to get the needed power as well, it is significantly more expensive to do so.
	// it would be better to cache the maximum needed power for a given resolution.
	getResolutionVoteInfo = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration, SUM(vr.power) AS approved_power
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.id = $1
	GROUP BY r.id, r.body, t.name;`

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
	getConfirmedResolutions = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration
	FROM ` + votingSchemaName + `.resolutions AS r
	INNER JOIN ` + votingSchemaName + `.resolution_types AS t ON r.type = t.id
	LEFT JOIN ` + votingSchemaName + `.votes AS v ON r.id = v.resolution_id
	LEFT JOIN ` + votingSchemaName + `.voters AS vr ON v.voter_id = vr.id
	WHERE r.body IS NOT NULL
	GROUP BY r.id, r.body, t.name
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

// SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration FROM resolutions AS r INNER JOIN resolution_types AS t ON r.type = t.id LEFT JOIN votes AS v ON r.id = v.resolution_id LEFT JOIN voters AS vr ON v.voter_id = vr.id WHERE r.body IS NOT NULL GROUP BY r.id HAVING SUM(vr.power) >= 3 ORDER BY r.id;

/* delete when we're sure we're done with the sprintf version of deleteResolutions

// formatResolutionList formats a list of resolutions for use in a sql statement
// it will hex encode the resolutions, and then wrap them in unhex()
func formatResolutionList(r []types.UUID) string {
	r = append(r, types.UUID{1, 2, 3})
	var buf strings.Builder
	for i, v := range r {
		buf.WriteString(`'\x`)
		buf.WriteString(hex.EncodeToString(v[:]))
		buf.WriteString(`'`)
		if i != len(r)-1 {
			buf.WriteString(",")
		}
	}

	return buf.String()
}
*/
