// The MIT License (MIT)

// Copyright (c) 2017 Fabian Wenzelmann

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sturavoting

import (
	"database/sql"
	"errors"
	"time"
)

const InvalidID = ^uint(0)

// DefaultTimeFromScanType is the default function to return database entries
// to a time.Time.
func TimeFromScanType(val interface{}) (time.Time, error) {
	// first check if we already got a time.Time because parseTime in
	// the MySQL driver is true
	if alreadyTime, ok := val.(time.Time); ok {
		return alreadyTime, nil
	}
	if bytes, ok := val.([]byte); ok {
		s := string(bytes)
		// let's hope this is correct... however who came up with THIS parse
		// function definition in Go?!
		return time.Parse("2006-01-02 15:04:05", s)
	} else {
		// we have to return some time... why not now.
		return time.Now().UTC(), errors.New("Invalid date in database, probably a bug if you end up here.")
	}
}

func initDB(db *sql.DB) error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS categories (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR(150) NOT NULL,
			created DATETIME NOT NULL,
			PRIMARY KEY (id),
			CONSTRAINT name_unique UNIQUE (name)
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS voters_revisions (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			category_id BIGINT UNSIGNED NOT NULL,
			created DATETIME NOT NULL,
			PRIMARY KEY (id),
			FOREIGN KEY (category_id)
				REFERENCES categories (id)
				ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS voters (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			revision_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(150),
			weight INT,
			PRIMARY KEY (id),
			CONSTRAINT name_unique UNIQUE (revision_id, name),
			FOREIGN KEY (revision_id)
				REFERENCES voters_revisions (id)
				ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS voting_collections (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			voters_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(150),
			voting_day DATETIME,
			PRIMARY KEY (id),
			CONSTRAINT name_unique UNIQUE (name),
			FOREIGN KEY (voters_id)
				REFERENCES voters_revisions (id)
				ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS voting_groups (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			collection_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(150),
			PRIMARY KEY (id),
			CONSTRAINT name_unique UNIQUE (collection_id, name),
			FOREIGN KEY (collection_id)
				REFERENCES voting_collections (id)
				ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS median_votings (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			group_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(150),
			max_value INT,
			percent_required DOUBLE,
			PRIMARY KEY (id),
			CONSTRAINT name_unique UNIQUE (group_id, name),
			FOREIGN KEY (group_id)
				REFERENCES voting_groups (id)
				ON DELETE CASCADE
		);
		`,
		// name should be unique in both tables
		// but we cannot enforce this
		`
		CREATE TABLE IF NOT EXISTS schulze_votings (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			group_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(150),
			percent_required DOUBLE,
			PRIMARY KEY (id),
			CONSTRAINT name_unique UNIQUE (group_id, name),
			FOREIGN KEY (group_id)
				REFERENCES voting_groups (id)
				ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS schulze_options (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			voting_id BIGINT UNSIGNED NOT NULL,
			option VARCHAR(150),
			PRIMARY KEY (id),
			CONSTRAINT option_unique UNIQUE (voting_id, option),
			FOREIGN KEY (voting_id)
				REFERENCES schulze_votings (id)
				ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS median_votes (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			voting_id BIGINT UNSIGNED NOT NULL,
			voter_id BIGINT UNSIGNED NOT NULL,
			value INT,
			PRIMARY KEY (id),
			CONSTRAINT vote_unique UNIQUE (voting_id, voter_id),
			FOREIGN KEY (voting_id)
				REFERENCES median_votings (id)
				ON DELETE CASCADE,
			FOREIGN KEY (voter_id)
				REFERENCES voters (id)
				ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS schulze_votes (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			option_id BIGINT UNSIGNED NOT NULL,
			voter_id BIGINT UNSIGNED NOT NULL,
			sorting_position INT,
			PRIMARY KEY (id),
			CONSTRAINT option_vote_unique UNIQUE (option_id, voter_id),
			FOREIGN KEY (option_id)
				REFERENCES schulze_options (id)
				ON DELETE CASCADE,
			FOREIGN KEY (voter_id)
				REFERENCES voters (id)
				ON DELETE CASCADE
		);
		`,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func InsertCategory(context *VotingContext, name string) error {
	now := Now()
	query := "INSERT INTO categories (name, created) VALUES (?, ?);"
	_, err := context.DB.Exec(query, name, now)
	return err
}

func ListCategories(context *VotingContext) ([]*Category, error) {
	query := "SELECT id, name, created FROM categories ORDER BY created"
	rows, err := context.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]*Category, 0)
	for rows.Next() {
		var id uint
		var name string
		var createdStr []byte
		scanErr := rows.Scan(&id, &name, &createdStr)
		if scanErr != nil {
			return nil, scanErr
		}
		created, timeErr := TimeFromScanType(createdStr)
		if timeErr != nil {
			return nil, timeErr
		}
		res = append(res, &Category{Name: name, ID: id, Created: created})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func InsertVotersRevision(context *VotingContext, categoryID uint) error {
	now := Now()
	query := "INSERT INTO voters_revisions (category_id, created) VALUES (?, ?);"
	_, err := context.DB.Exec(query, categoryID, now)
	return err
}

func ListVotersRevision(context *VotingContext, categoryID uint) ([]*VotersRevision, error) {
	query := "SELECT id, category_id, created FROM voters_revisions ORDER BY created"
	args := make([]interface{}, 0)
	if categoryID != InvalidID {
		query = "SELECT id, category_id, created FROM voters_revisions WHERE category_id = ? ORDER BY created"
		args = append(args, categoryID)
	}
	rows, err := context.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]*VotersRevision, 0)
	for rows.Next() {
		var id, cID uint
		var createdStr []byte
		scanErr := rows.Scan(&id, &cID, &createdStr)
		if scanErr != nil {
			return nil, scanErr
		}
		created, timeErr := TimeFromScanType(createdStr)
		if timeErr != nil {
			return nil, timeErr
		}
		res = append(res, &VotersRevision{CategoryID: cID, ID: id, Created: created})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func InsertVoters(context *VotingContext, revisionID uint, voters []*Voter) error {
	tx, err := context.DB.Begin()
	if err != nil {
		return err
	}
	query := "INSERT INTO voters (revision_id, name, weight) VALUES (?, ?, ?);"
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	err = nil
	for _, voter := range voters {
		_, err = stmt.Exec(revisionID, voter.Name, voter.Weight)
		if err != nil {
			break
		}
	}
	if err == nil {
		return tx.Commit()
	} else {
		rollBackErr := tx.Rollback()
		if rollBackErr != nil {
			context.Logger.WithError(rollBackErr).Error("Error while using Rollback in InsertVotersRevision")
		}
		return err
	}
}

func ListVoters(context *VotingContext, revisionID uint) ([]*Voter, error) {
	query := "SELECT id, revision_id, name, weight FROM voters ORDER BY name"
	args := make([]interface{}, 0)
	if revisionID != InvalidID {
		query = "SELECT id, revision_id, name, weight FROM voters WHERE revision_id = ? ORDER BY name"
		args = append(args, revisionID)
	}
	rows, err := context.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]*Voter, 0)
	for rows.Next() {
		var id, rID uint
		var name string
		var weight int
		scanErr := rows.Scan(&id, &rID, &name, &weight)
		if scanErr != nil {
			return nil, scanErr
		}
		res = append(res, &Voter{ID: id, RevisionID: rID, Name: name, Weight: weight})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}
