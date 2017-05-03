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
)

const InvalidID = ^uint(0)

func initDB(db *sql.DB) error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS categories (
			id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(150) NOT NULL,
			created DATETIME NOT NULL,
			PRIMARY KEY (id)
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS voters_revisions (
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
			id BIGINT UNSIGNED NOT NULL,
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
