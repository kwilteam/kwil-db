package voting

import (
	"bytes"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/core/types"
)

/*
	sql.go contains information about the sql tables used by the voting system.
*/

// Datastore is a connection to a database.

var (
	// tableResolutions is the sql table used to store resolutions that can be voted on
	tableResolutions = `CREATE TABLE IF NOT EXISTS resolutions (
		id BLOB PRIMARY KEY, -- id is an rfc4122 uuid derived from the body
		body BLOB, -- body is the actual resolution info
		type BLOB, -- type is the type of resolution
		expiration INTEGER NOT NULL, -- expiration is the blockheight at which the resolution expires
		FOREIGN KEY(type) REFERENCES resolution_types(id) ON UPDATE CASCADE ON DELETE CASCADE
	) WITHOUT ROWID, STRICT;`

	// resolutionTypeIndex is the sql index used to index the type of a resolution
	resolutionTypeIndex = `CREATE INDEX IF NOT EXISTS type_index ON resolutions (type);`

	tableResolutionTypes = `CREATE TABLE IF NOT EXISTS resolution_types (
		id BLOB PRIMARY KEY, -- id is an rfc4122 uuid derived from the name
		name TEXT UNIQUE NOT NULL -- name is the name of the resolution type
	) WITHOUT ROWID, STRICT;`

	tableVoters = `CREATE TABLE IF NOT EXISTS voters (
		id BLOB PRIMARY KEY, -- id is an rfc4122 uuid derived from the voter
		name BLOB UNIQUE NOT NULL, -- voter is the identifier of the voter
		power INTEGER NOT NULL CHECK(power > 0) -- power is the voting power of the voter
	) WITHOUT ROWID, STRICT;`

	// votes tracks whether a voter has voted on a resolution
	tableVotes = `CREATE TABLE IF NOT EXISTS votes (
		resolution_id BLOB NOT NULL, 
		voter_id BLOB NOT NULL,
		FOREIGN KEY(resolution_id) REFERENCES resolutions(id) ON UPDATE CASCADE ON DELETE CASCADE,
		FOREIGN KEY(voter_id) REFERENCES voters(id) ON UPDATE CASCADE ON DELETE CASCADE,
		PRIMARY KEY(resolution_id, voter_id)
	) WITHOUT ROWID, STRICT;`

	// creating a votesResolutionIndex since looking up votes by resolution is a common operation
	votesResolutionIndex = `CREATE INDEX IF NOT EXISTS resolution_index ON votes (resolution_id);`
	// we don't look up votes by voter, so we don't need a voter index

	// tableProcessed contains all processed resolution ids
	tableProcessed = `CREATE TABLE IF NOT EXISTS processed (
		id BLOB PRIMARY KEY
	) WITHOUT ROWID, STRICT;`

	// resolutionIDExists is the sql statement used to ensure a resolution ID is present in the resolutions table
	resolutionIDExists = `INSERT INTO resolutions (id, expiration) VALUES ($id, $expiration) ON CONFLICT(id) DO NOTHING;`

	// upsertResolution is the sql statement used to ensure a resolution is present in the resolutions table
	upsertResolution = `INSERT INTO resolutions (id, body, type, expiration)
	VALUES ($id, $body, (
		SELECT id
		FROM resolution_types
		WHERE name = $type
	), $expiration)
	ON CONFLICT(id)
		DO UPDATE
		SET body = $body, type = (
			SELECT id
			FROM resolution_types
			WHERE name = $type
		),
		expiration = $expiration;`

	// upsertVoter is the sql statement used to ensure a voter is present in the voters table.  If the voter is present, the power is updated.
	upsertVoter = `INSERT INTO voters (id, name, power) VALUES ($id, $voter, $power) ON CONFLICT(id) DO UPDATE SET power = power + $power;`

	// decreaseVoterPower is the sql statement used to decrease the power of a voter
	// this is necessary because the voters table CHECK filters before the on conflict
	decreaseVoterPower = `UPDATE voters SET power = power - $power WHERE id = $id;`

	// removeVoter is the sql statement used to remove a voter from the voters table
	removeVoter = `DELETE FROM voters WHERE id = $id;`

	// getVoterPower is the sql statement used to get the power and name of a voter
	getVoterPower = `SELECT power FROM voters WHERE id = $id;`

	// addVote adds a vote for a resolution
	addVote = `INSERT INTO votes (resolution_id, voter_id) VALUES ($resolution_id, $voter_id) ON CONFLICT(resolution_id, voter_id) DO NOTHING;`

	// hasVoted checks if a voter has voted on a resolution
	hasVoted = `SELECT resolution_id FROM votes WHERE resolution_id = $resolution_id AND voter_id = $voter_id;`

	// expireResolutions is the sql statement used to expire resolutions
	// it will expire resolutions that have an expiration less than or equal to the given blockheight
	expireResolutions = `DELETE FROM resolutions WHERE expiration <= $blockheight RETURNING id;`

	// getResolutionBody gets a resolution body by id
	getResolutionBody = `SELECT body FROM resolutions WHERE id = $id;`

	// getResolutionVoteInfo is the sql statement used to get a resolution and the associated vote info.
	// while it would be nice to get the needed power as well, it is significantly more expensive to do so.
	// it would be better to cache the maximum needed power for a given resolution.
	getResolutionVoteInfo = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration, SUM(vr.power) AS approved_power
	FROM resolutions AS r
	INNER JOIN resolution_types AS t ON r.type = t.id
	LEFT JOIN votes AS v ON r.id = v.resolution_id
	LEFT JOIN voters AS vr ON v.voter_id = vr.id
	WHERE r.id = $id
	GROUP BY r.id;`

	// getUnfilledResolutionVoteInfo gets en expiration and approved power for a resolution that has not been filled with a body and type
	getUnfilledResolutionVoteInfo = `SELECT r.expiration AS expiration, SUM(vr.power) AS approved_power
	FROM resolutions AS r
	LEFT JOIN votes AS v ON r.id = v.resolution_id
	LEFT JOIN voters AS vr ON v.voter_id = vr.id
	WHERE r.id = $id
	GROUP BY r.id;`

	// resolutionsByType is the sql statement used to get resolutions of a given type
	resolutionsByType = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration
	FROM resolutions AS r
	INNER JOIN resolution_types AS t ON r.type = t.id
	WHERE t.name = $type;`

	// getConfirmedResolutions is the statement used to get all resolutions
	// that have been confirmed above the given power threshold
	// we do not calculate the threshold here since we need to guarantee accuracy
	// using big ints.
	// it orders by id for determinism
	getConfirmedResolutions = `SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration
	FROM resolutions AS r
	INNER JOIN resolution_types AS t ON r.type = t.id
	LEFT JOIN votes AS v ON r.id = v.resolution_id
	LEFT JOIN voters AS vr ON v.voter_id = vr.id
	WHERE r.body IS NOT NULL
	GROUP BY r.id
	HAVING SUM(vr.power) >= $power_needed
	ORDER BY r.id;`

	// deleteResolutions deletes a set of resolutions
	// it is meant to be used in formatResolutionList
	deleteResolutions = `DELETE FROM resolutions WHERE id IN (%s);`

	// totalPower gets the total power of all voters
	totalPower = `SELECT SUM(power) AS required_power FROM voters;`

	// createResolutionType creates a resolution type
	createResolutionType = `INSERT INTO resolution_types (id, name) VALUES ($id, $name) ON CONFLICT(id) DO NOTHING;`

	// markProcessed marks a resolution as processed
	markProcessed = `INSERT INTO processed (id) VALUES ($id);`

	// alreadyProcessed checks if a resolution has already been processed
	alreadyProcessed = `SELECT id FROM processed WHERE id = $id;`
)

// SELECT r.id AS id, r.body AS body, t.name AS type, r.expiration AS expiration FROM resolutions AS r INNER JOIN resolution_types AS t ON r.type = t.id LEFT JOIN votes AS v ON r.id = v.resolution_id LEFT JOIN voters AS vr ON v.voter_id = vr.id WHERE r.body IS NOT NULL GROUP BY r.id HAVING SUM(vr.power) >= 3 ORDER BY r.id;

// formatResolutionList formats a list of resolutions for use in a sql statement
// it will hex encode the resolutions, and then wrap them in unhex()
func formatResolutionList(r []types.UUID) string {
	var buf bytes.Buffer
	for i, v := range r {
		buf.WriteString("unhex('")
		buf.WriteString(hex.EncodeToString(v[:]))
		buf.WriteString("')")
		if i != len(r)-1 {
			buf.WriteString(", ")
		}
	}

	return buf.String()
}
