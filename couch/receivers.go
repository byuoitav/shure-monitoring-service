package couch

import (
	"context"
	"fmt"

	"github.com/byuoitav/shure-monitoring-service"
)

const _receiversDB = "shure-receivers"
const _receiversDefaultDoc = "default"

type receiverDoc struct {
	ID        string     `json:"_id"`
	Receivers []receiver `json:"receivers"`
}

type receiver struct {
	Name    string `json:"name"`
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
