package sync

import (
	"sync"

	"compass/logger"
	"compass/namerd"

	"github.com/rs/zerolog"
)

type Syncer interface {
	Sync(dtab namerd.Dtab)
}

type Sync struct {
	client *namerd.Client
	log    zerolog.Logger
	// channels
	syncC chan namerd.Dtab
	stopC chan bool
	// Wait Group
	wg sync.WaitGroup
}

func (s *Sync) Start() {
	s.wg.Add(1)
	defer s.wg.Done()
	s.log.Debug().Msg("starting dtab sync loop")
	for {
		select {
		case <-s.stopC:
			// TODO: exit - closing
			return
		case dtab := <-s.syncC:
			// TODO: sync
			s.log.Debug().Str("dtab", dtab.String()).Msg("sync dtab")
		}
	}
}

func (s *Sync) Sync(d namerd.Dtab) {
	select {
	case <-s.stopC:
		// TODO: log
		return
	default:
		s.syncC <- d
	}
}

func (s *Sync) Stop() {
	defer s.log.Debug().Msg("stopped syncing dtabs")
	close(s.stopC)
	s.wg.Wait()
}

func New(nc *namerd.Client) *Sync {
	return &Sync{
		client: nc,
		log:    logger.New(),
		syncC:  make(chan namerd.Dtab),
		stopC:  make(chan bool),
	}
}
