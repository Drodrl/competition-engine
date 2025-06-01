package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
)

type entrant struct{ UserID, TeamID *int }

// GenerateRoundRobin will insert N–1 rounds and all their matches & participants.
// Assumes an even number of entries in stage_participants.
func GenerateRoundRobin(db *sql.DB, stageID int) error {

	rows, err := db.Query(
		`SELECT user_id, team_id FROM stage_participants WHERE stage_id=$1`,
		stageID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var entrants []entrant
	for rows.Next() {
		var e entrant
		if err := rows.Scan(&e.UserID, &e.TeamID); err != nil {
			return fmt.Errorf("failed to scan participant: %w", err)
		}
		entrants = append(entrants, e)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("row error: %w", err)
	}

	N := len(entrants)
	if N == 0 {
		return fmt.Errorf("no participants in stage")
	}
	if N%2 != 0 {
		return fmt.Errorf("expected even participants, got %d", N)
	}
	rounds := N - 1
	log.Printf("Number of rounds: %d ", rounds)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// insert rounds
	roundIDs := make([]int, rounds)
	for i := 1; i <= rounds; i++ {
		if err := tx.QueryRow(
			`INSERT INTO rounds (stage_id, round_number) VALUES ($1,$2) RETURNING round_id`,
			stageID, i,
		).Scan(&roundIDs[i-1]); err != nil {
			return fmt.Errorf("failed to insert round: %w", err)
		}
	}

	// build circle
	idx := make([]int, N)
	for i := range idx {
		idx[i] = i
	}

	for _, rid := range roundIDs {
		for i := 0; i < N/2; i++ {
			a, b := entrants[idx[i]], entrants[idx[N-1-i]]
			var mid int
			if err := tx.QueryRow(
				`INSERT INTO matches (round_id, scheduled_at) VALUES ($1, NOW()) RETURNING match_id`,
				rid,
			).Scan(&mid); err != nil {
				return fmt.Errorf("failed to insert match: %w", err)
			}

			_, err := tx.Exec(
				`INSERT INTO match_participants
                        (match_id,user_id,team_id,is_winner,score)
                       VALUES
                        ($1,$2,$3,false,NULL),
                        ($1,$4,$5,false,NULL)`,
				mid,
				a.UserID, a.TeamID,
				b.UserID, b.TeamID,
			)
			if err != nil {
				return fmt.Errorf("failed to insert match participants: %w", err)
			}
		}
		// rotate (keep 0 fixed)
		tmp := idx[1]
		copy(idx[1:], idx[2:])
		idx[N-1] = tmp
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GenerateRoundSingleElim(db *sql.DB, stageID int) error {

	// 1. Find the next round number
	var nextRound int
	err := db.QueryRow(
		`SELECT COALESCE(MAX(round_number), 0) + 1 FROM rounds WHERE stage_id = $1`,
		stageID,
	).Scan(&nextRound)
	if err != nil {
		return fmt.Errorf("failed to get next round number: %w", err)
	}

	// 2. Fetch participants for this round
	var rows *sql.Rows
	var entrants []entrant
	if nextRound == 1 {
		rows, err = db.Query(
			`SELECT user_id, team_id FROM stage_participants WHERE stage_id=$1`,
			stageID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var e entrant
			rows.Scan(&e.UserID, &e.TeamID)
			entrants = append(entrants, e)
		}
	} else {
		// Later rounds: only winners from previous round
		rows, err = db.Query(
			`SELECT mp.user_id, mp.team_id
			 FROM match_participants mp
			 JOIN matches m ON mp.match_id = m.match_id
			 JOIN rounds r ON m.round_id = r.round_id
			 WHERE r.stage_id = $1 AND r.round_number = $2 AND mp.is_winner = true`,
			stageID, nextRound-1,
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var e entrant
			rows.Scan(&e.UserID, &e.TeamID)
			entrants = append(entrants, e)
		}
	}
	N := len(entrants)
	if N%2 != 0 {
		return fmt.Errorf("expected even participants, got %d", N)
	}

	tx, _ := db.Begin()
	var roundID int
	tx.QueryRow(
		`INSERT INTO rounds (stage_id, round_number) VALUES ($1,$2) RETURNING round_id`,
		stageID, nextRound,
	).Scan(&roundID)

	// Insert matches for this round
	for i := 0; i < N; i += 2 {
		a := entrants[i]
		b := entrants[i+1]
		var matchID int
		err := tx.QueryRow(
			`INSERT INTO matches (round_id, scheduled_at) VALUES ($1, NOW()) RETURNING match_id`,
			roundID,
		).Scan(&matchID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert match: %w", err)
		}

		// Insert both participants into match_participants
		_, err = tx.Exec(
			`INSERT INTO match_participants (match_id, user_id, team_id, is_winner, score)
			 VALUES ($1, $2, $3, false, NULL), ($1, $4, $5, false, NULL)`,
			matchID,
			a.UserID, a.TeamID,
			b.UserID, b.TeamID,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert match participants: %w", err)
		}
	}

	return tx.Commit()
}

func GenerateRoundDoubleElim(db *sql.DB, stageID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 1. Find the next round number for each bracket
	var nextWinnersRound, nextLosersRound int
	err = tx.QueryRow(
		`SELECT COALESCE(MAX(round_number), 0) + 1 FROM rounds WHERE stage_id = $1 AND bracket = 'W'`,
		stageID,
	).Scan(&nextWinnersRound)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get next winners round: %w", err)
	}
	err = tx.QueryRow(
		`SELECT COALESCE(MAX(round_number), 0) + 1 FROM rounds WHERE stage_id = $1 AND bracket = 'L'`,
		stageID,
	).Scan(&nextLosersRound)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get next losers round: %w", err)
	}

	// 2. Fetch participants for the next winners round
	var winners []entrant
	if nextWinnersRound == 1 {
		winnersRows, err := tx.Query(
			`SELECT user_id, team_id FROM stage_participants WHERE stage_id=$1`,
			stageID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer winnersRows.Close()
		for winnersRows.Next() {
			var e entrant
			winnersRows.Scan(&e.UserID, &e.TeamID)
			winners = append(winners, e)
		}
	} else {
		// Later rounds: only winners from previous winners round
		winnersRows, err := tx.Query(
			`SELECT mp.user_id, mp.team_id
             FROM match_participants mp
             JOIN matches m ON mp.match_id = m.match_id
             JOIN rounds r ON m.round_id = r.round_id
             WHERE r.stage_id = $1 AND r.bracket = 'W' AND r.round_number = $2 AND mp.is_winner = true`,
			stageID, nextWinnersRound-1,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer winnersRows.Close()
		for winnersRows.Next() {
			var e entrant
			winnersRows.Scan(&e.UserID, &e.TeamID)
			winners = append(winners, e)
		}
	}
	Nw := len(winners)

	// 3. Fetch participants for the next losers round (skip if first round)
	var losers []entrant
	var Nl int
	if nextWinnersRound > 1 {
		var losersRows *sql.Rows
		losersRows, err = tx.Query(
			`SELECT mp.user_id, mp.team_id
             FROM match_participants mp
             JOIN matches m ON mp.match_id = m.match_id
             JOIN rounds r ON m.round_id = r.round_id
             WHERE r.stage_id = $1 AND (
                    (r.bracket = 'W' AND r.round_number = $2 AND mp.is_winner = false)
                    OR
                    (r.bracket = 'L' AND r.round_number = $3 AND mp.is_winner = true)
             )`,
			stageID, nextWinnersRound-1, nextLosersRound-1,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer losersRows.Close()

		for losersRows.Next() {
			var e entrant
			losersRows.Scan(&e.UserID, &e.TeamID)
			losers = append(losers, e)
		}
		Nl = len(losers)
	}

	// 4. Check for grand final condition: 1 winner and 1 loser remain
	if Nw == 1 && Nl == 1 {
		// Insert grand final round
		var grandFinalRoundID int
		err = tx.QueryRow(
			`INSERT INTO rounds (stage_id, round_number, bracket) VALUES ($1, 1, 'G') RETURNING round_id`,
			stageID,
		).Scan(&grandFinalRoundID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert grand final round: %w", err)
		}
		var matchID int
		err = tx.QueryRow(
			`INSERT INTO matches (round_id, scheduled_at) VALUES ($1, NOW()) RETURNING match_id`,
			grandFinalRoundID,
		).Scan(&matchID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert grand final match: %w", err)
		}
		_, err = tx.Exec(
			`INSERT INTO match_participants (match_id, user_id, team_id, is_winner, score)
             VALUES ($1, $2, $3, false, NULL), ($1, $4, $5, false, NULL)`,
			matchID,
			winners[0].UserID, winners[0].TeamID,
			losers[0].UserID, losers[0].TeamID,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert grand final participants: %w", err)
		}
		return tx.Commit()
	}

	// 5. Insert next winners round (if more than 1 in winners bracket)
	if Nw > 1 {
		if Nw%2 != 0 {
			tx.Rollback()
			return fmt.Errorf("expected even participants in winners bracket, got %d", Nw)
		}
		var winnersRoundID int
		err = tx.QueryRow(
			`INSERT INTO rounds (stage_id, round_number, bracket) VALUES ($1, $2, 'W') RETURNING round_id`,
			stageID, nextWinnersRound,
		).Scan(&winnersRoundID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert winners round: %w", err)
		}
		for i := 0; i < Nw; i += 2 {
			a := winners[i]
			b := winners[i+1]
			var matchID int
			err := tx.QueryRow(
				`INSERT INTO matches (round_id, scheduled_at) VALUES ($1, NOW()) RETURNING match_id`,
				winnersRoundID,
			).Scan(&matchID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert winners match: %w", err)
			}
			_, err = tx.Exec(
				`INSERT INTO match_participants (match_id, user_id, team_id, is_winner, score)
                 VALUES ($1, $2, $3, false, NULL), ($1, $4, $5, false, NULL)`,
				matchID,
				a.UserID, a.TeamID,
				b.UserID, b.TeamID,
			)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert winners match participants: %w", err)
			}
		}
	}

	// 6. Insert next losers round (if more than 1 in losers bracket)
	if nextWinnersRound > 1 && Nl > 0 {
		var losersRoundID int
		err = tx.QueryRow(
			`INSERT INTO rounds (stage_id, round_number, bracket) VALUES ($1, $2, 'L') RETURNING round_id`,
			stageID, nextLosersRound,
		).Scan(&losersRoundID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert losers round: %w", err)
		}
		if Nl == 2 {
			// Losers bracket final: only one match
			a := losers[0]
			b := losers[1]
			var matchID int
			err := tx.QueryRow(
				`INSERT INTO matches (round_id, scheduled_at) VALUES ($1, NOW()) RETURNING match_id`,
				losersRoundID,
			).Scan(&matchID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert losers final match: %w", err)
			}
			_, err = tx.Exec(
				`INSERT INTO match_participants (match_id, user_id, team_id, is_winner, score)
                 VALUES ($1, $2, $3, false, NULL), ($1, $4, $5, false, NULL)`,
				matchID,
				a.UserID, a.TeamID,
				b.UserID, b.TeamID,
			)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert losers final match participants: %w", err)
			}
		} else if Nl > 2 {
			if Nl%2 != 0 {
				tx.Rollback()
				return fmt.Errorf("expected even participants in losers bracket, got %d", Nl)
			}
			for i := 0; i < Nl; i += 2 {
				a := losers[i]
				b := losers[i+1]
				var matchID int
				err := tx.QueryRow(
					`INSERT INTO matches (round_id, scheduled_at) VALUES ($1, NOW()) RETURNING match_id`,
					losersRoundID,
				).Scan(&matchID)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to insert losers match: %w", err)
				}
				_, err = tx.Exec(
					`INSERT INTO match_participants (match_id, user_id, team_id, is_winner, score)
                     VALUES ($1, $2, $3, false, NULL), ($1, $4, $5, false, NULL)`,
					matchID,
					a.UserID, a.TeamID,
					b.UserID, b.TeamID,
				)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to insert losers match participants: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// Returns a slice of entrants (user_id, team_id) for the top N participants in the previous round robin stage.
func GetTopNFromPrevRoundRobin(db *sql.DB, currentStageID int, n int) ([]entrant, error) {
	// 1. Find previous stage (by stage_order)
	var prevStageID int
	err := db.QueryRow(`
        SELECT stage_id FROM competition_stages
        WHERE stage_order = (
            SELECT stage_order - 1 FROM competition_stages WHERE stage_id = $1
        ) AND competition_id = (
            SELECT competition_id FROM competition_stages WHERE stage_id = $1
        )
    `, currentStageID).Scan(&prevStageID)
	if err != nil {
		return nil, fmt.Errorf("could not find previous stage: %w", err)
	}

	// 2. Check if previous stage is round robin
	var prevFormatID int
	err = db.QueryRow(`SELECT tourney_format_id FROM competition_stages WHERE stage_id = $1`, prevStageID).Scan(&prevFormatID)
	if err != nil {
		return nil, fmt.Errorf("could not get previous stage format: %w", err)
	}
	if prevFormatID != 3 { // 3 = RoundRobin
		return nil, fmt.Errorf("previous stage is not round robin")
	}

	// 3. Get all participants in previous stage
	type participant struct {
		UserID *int
		TeamID *int
	}
	var participants []participant
	rows, err := db.Query(`SELECT user_id, team_id FROM stage_participants WHERE stage_id = $1`, prevStageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p participant
		if err := rows.Scan(&p.UserID, &p.TeamID); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}

	// 4. Calculate wins for each participant
	type score struct {
		Entrant participant
		Wins    int
	}
	scores := make([]score, 0, len(participants))
	for _, p := range participants {
		var wins int
		if p.UserID != nil {
			db.QueryRow(`
                SELECT COUNT(*) FROM match_participants mp
                JOIN matches m ON mp.match_id = m.match_id
                JOIN rounds r ON m.round_id = r.round_id
                WHERE r.stage_id = $1 AND mp.user_id = $2 AND mp.is_winner = true
            `, prevStageID, *p.UserID).Scan(&wins)
		} else if p.TeamID != nil {
			db.QueryRow(`
                SELECT COUNT(*) FROM match_participants mp
                JOIN matches m ON mp.match_id = m.match_id
                JOIN rounds r ON m.round_id = r.round_id
                WHERE r.stage_id = $1 AND mp.team_id = $2 AND mp.is_winner = true
            `, prevStageID, *p.TeamID).Scan(&wins)
		}
		scores = append(scores, score{Entrant: p, Wins: wins})
	}

	// 5. Sort by wins descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Wins > scores[j].Wins
	})

	// 6. Head-to-head tiebreaker for ties
	// For each tie group, check head-to-head result
	i := 0
	for i < len(scores) {
		j := i + 1
		for j < len(scores) && scores[j].Wins == scores[i].Wins {
			j++
		}
		if j-i > 1 {
			// Tie group from i to j-1
			sort.SliceStable(scores[i:j], func(a, b int) bool {
				pa := scores[i+a].Entrant
				pb := scores[i+b].Entrant
				var headToHeadWinsA, headToHeadWinsB int
				if pa.UserID != nil && pb.UserID != nil {
					db.QueryRow(`
                        SELECT COUNT(*) FROM matches m
                        JOIN rounds r ON m.round_id = r.round_id
                        JOIN match_participants mp1 ON mp1.match_id = m.match_id AND mp1.user_id = $1
                        JOIN match_participants mp2 ON mp2.match_id = m.match_id AND mp2.user_id = $2
                        WHERE r.stage_id = $3 AND mp1.is_winner = true
                    `, *pa.UserID, *pb.UserID, prevStageID).Scan(&headToHeadWinsA)
					db.QueryRow(`
                        SELECT COUNT(*) FROM matches m
                        JOIN rounds r ON m.round_id = r.round_id
                        JOIN match_participants mp1 ON mp1.match_id = m.match_id AND mp1.user_id = $1
                        JOIN match_participants mp2 ON mp2.match_id = m.match_id AND mp2.user_id = $2
                        WHERE r.stage_id = $3 AND mp1.is_winner = true
                    `, *pb.UserID, *pa.UserID, prevStageID).Scan(&headToHeadWinsB)
				} else if pa.TeamID != nil && pb.TeamID != nil {
					db.QueryRow(`
                        SELECT COUNT(*) FROM matches m
                        JOIN rounds r ON m.round_id = r.round_id
                        JOIN match_participants mp1 ON mp1.match_id = m.match_id AND mp1.team_id = $1
                        JOIN match_participants mp2 ON mp2.match_id = m.match_id AND mp2.team_id = $2
                        WHERE r.stage_id = $3 AND mp1.is_winner = true
                    `, *pa.TeamID, *pb.TeamID, prevStageID).Scan(&headToHeadWinsA)
					db.QueryRow(`
                        SELECT COUNT(*) FROM matches m
                        JOIN rounds r ON m.round_id = r.round_id
                        JOIN match_participants mp1 ON mp1.match_id = m.match_id AND mp1.team_id = $1
                        JOIN match_participants mp2 ON mp2.match_id = m.match_id AND mp2.team_id = $2
                        WHERE r.stage_id = $3 AND mp1.is_winner = true
                    `, *pb.TeamID, *pa.TeamID, prevStageID).Scan(&headToHeadWinsB)
				}
				return headToHeadWinsA > headToHeadWinsB
			})
		}
		i = j
	}

	// 7. Return top N
	top := make([]entrant, 0, n)
	for i := 0; i < n && i < len(scores); i++ {
		top = append(top, entrant{UserID: scores[i].Entrant.UserID, TeamID: scores[i].Entrant.TeamID})
	}
	return top, nil
}
