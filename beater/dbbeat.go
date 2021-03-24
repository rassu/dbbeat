package beater

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/ronaudinho/dbbeat/config"
	"github.com/ronaudinho/dbbeat/postgres"
)

// dbBeat configuration.
type dbBeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
	db     *postgres.Conn
}

// New creates an instance of dbBeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	log := logp.NewLogger("config")
	log.Debugf("config", "%s", c.DBConfig.URI)

	bt := &dbBeat{
		done:   make(chan struct{}),
		config: c,
		db:     postgres.NewConn(c.DBConfig),
	}
	return bt, nil
}

// Run starts dbBeat.
func (bt *dbBeat) Run(b *beat.Beat) error {
	log := logp.NewLogger("run")

	log.Info("dbBeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	err = bt.db.Connect()
	if err != nil {
		return err
	}
	err = bt.db.Listen()
	if err != nil {
		return err
	}

	for {
		select {
		case <-bt.done:
			return nil
		case n := <-bt.db.Listener.Notify:
			chOps := strings.Split(n.Channel, "_")
			tableName := strings.Join(chOps[0:len(chOps)-1], "_")
			var payload map[string]interface{}
			err := json.Unmarshal([]byte(n.Extra), &payload)
			if err != nil {
				log.Info(err.Error())
			}
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"channel": tableName,
					"ops":     chOps[len(chOps)-1],
					"payload": payload,
				},
			}

			if event.Meta == nil {
				event.Meta = make(map[string]interface{})
			}

			event.Meta["raw_index"] = fmt.Sprintf("%v-%v", bt.db.DBName, tableName)

			bt.client.Publish(event)
			log.Info("Event sent")
		}
	}
}

// Stop stops dbBeat.
func (bt *dbBeat) Stop() {
	bt.client.Close()
	bt.db.Close()
	close(bt.done)
}
