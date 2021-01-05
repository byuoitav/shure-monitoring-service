package couch

import (
	"context"
	"fmt"

	"github.com/byuoitav/shure-monitoring-service"
)

const _receiversDB = "shure-receivers"
const _receiversDefaultDoc = "default"
const _devicesDB = "devices"

type receiverDoc struct {
	ID        string     `json:"_id"`
	Receivers []receiver `json:"receivers"`
}

type receiver struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// device is the couch representation of a device
type device struct {
	ID      string `json:"_id"`
	Address string `json:"address"`
}

func (s *Service) GetReceivers() ([]shure.Receiver, error) {
	doc := receiverDoc{}

	err := s.client.DB(context.TODO(), _receiversDB).
		Get(context.TODO(), _receiversDefaultDoc).
		ScanDoc(&doc)
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving receivers: %s", err)
	}

	receivers := []shure.Receiver{}

	for _, r := range doc.Receivers {
		receivers = append(receivers, shure.Receiver{
			Name:    r.Name,
			Address: r.Address,
		})
	}

	return receivers, nil
}

func (s *Service) GetRoomReceivers(roomID string) ([]shure.Receiver, error) {
	db := s.client.DB(context.TODO(), _devicesDB)

	// Query
	q := query{
		Selector: map[string]interface{}{
			"_id": search{
				Regex: fmt.Sprintf("%s-RCV", roomID),
			},
		},
		Limit: 1000,
	}

	// Make the request
	rows, err := db.Find(context.TODO(), q)
	if err != nil {
		return nil, fmt.Errorf("couch/GetRoomReceivers couch request: %w", err)
	}

	recvs := []shure.Receiver{}
	for rows.Next() {
		r := device{}
		err := rows.ScanDoc(&r)
		if err != nil {
			return nil, fmt.Errorf("couch/GetRoomReceivers unmarshal: %w", err)
		}
		// Convert
		rec := shure.Receiver{
			Name:    r.ID,
			Address: r.Address,
		}
		recvs = append(recvs, rec)
	}

	return recvs, nil

}
