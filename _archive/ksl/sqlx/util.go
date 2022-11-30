package sqlx

import (
	"database/sql"
	"fmt"
	"hash/crc32"
	"strings"
)

const advisoryLockIDSalt uint = 1486364155

func GenerateAdvisoryLockId(name string, additionalNames ...string) (string, error) {
	if len(additionalNames) > 0 {
		name = strings.Join(append(additionalNames, name), "\x00")
	}
	sum := crc32.ChecksumIEEE([]byte(name))
	sum = sum * uint32(advisoryLockIDSalt)
	return fmt.Sprint(sum), nil
}

// ScanOne scans one record and closes the rows at the end.
func ScanOne(rows *sql.Rows, dest ...any) error {
	defer rows.Close()
	if !rows.Next() {
		return sql.ErrNoRows
	}
	if err := rows.Scan(dest...); err != nil {
		return err
	}
	return rows.Close()
}

// ScanNullBool scans one sql.NullBool record and closes the rows at the end.
func ScanNullBool(rows *sql.Rows) (sql.NullBool, error) {
	var b sql.NullBool
	return b, ScanOne(rows, &b)
}
