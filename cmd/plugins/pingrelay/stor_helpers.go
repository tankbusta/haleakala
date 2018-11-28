package pingrelay

import (
	"github.com/asdine/storm"
)

var DefaultMaxSimilarity float64 = 0.75

type pingRecord struct {
	ID                   int    // primary key
	Nickname             string `storm:"index"`
	MaxMessageSimilarity *float64
	Blocked              *bool
}

func (s *zncRelayPlugin) getPingRecord(user string) (*pingRecord, error) {
	var rec pingRecord

	txn, err := s.stor.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Commit()

	if err := txn.One("Nickname", user, &rec); err != nil {
		if err == storm.ErrNotFound {
			goto CreateRecord
		}
	}

	return &rec, nil

CreateRecord:
	rec = pingRecord{
		Nickname:             user,
		MaxMessageSimilarity: &DefaultMaxSimilarity,
	}

	if err := txn.Save(rec); err != nil {
		return &rec, err
	}

	return &rec, err
}

func (s *zncRelayPlugin) updatePingRecord(rec *pingRecord) error {
	txn, err := s.stor.Begin(true)
	if err != nil {
		return err
	}
	defer txn.Commit()

	if err := txn.One("Nickname", rec.Nickname, &rec); err != nil {
		return err
	}

	return txn.Update(rec)
}
