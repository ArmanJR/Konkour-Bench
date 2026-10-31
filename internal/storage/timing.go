package storage

// AttemptTiming excludes responses and request bodies. Reports use the original
// inference attempt, including when a later run reused its cached answer.
type AttemptTiming struct {
	ID                       int64
	RunID, Started, Finished string
}

func (s *Store) AttemptTimings() (map[int64]AttemptTiming, error) {
	rows, err := s.db.Query("SELECT id,run_id,started,finished FROM attempts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[int64]AttemptTiming{}
	for rows.Next() {
		var a AttemptTiming
		if err := rows.Scan(&a.ID, &a.RunID, &a.Started, &a.Finished); err != nil {
			return nil, err
		}
		result[a.ID] = a
	}
	return result, rows.Err()
}
