package sync

import (
	"context"

	"compass/namerd"
	"compass/store"

	"github.com/rs/zerolog"
)

// A DtabUpdatePublisher publishes delegation table names to a channel
// when the should be synced with namerd
type DtabUpdatePublisher interface {
	DtabUpdates() <-chan namerd.Dtab
}

type Sync struct {
	namerd    *namerd.Client
	store     store.DentriesByDtabSelector
	publisher DtabUpdatePublisher
	log       zerolog.Logger
}

func (s *Sync) Start(ctx context.Context) {
	log := zerolog.Ctx(ctx)
	log.Debug().Msg("starting dtab sync loop")
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("stopped dtab sync loop")
			return
		case dtab := <-s.publisher.DtabUpdates():
			log.Debug().Str("dtab", dtab.String()).Msg("sync dtab")
			if err := syncDtab(ctx, s.store, s.namerd, dtab); err != nil {
				log.Error().Err(err).Msg("error syncing dtab dentries")
			}
		}
	}
}

func New(n *namerd.Client, s store.DentriesByDtabSelector, p DtabUpdatePublisher) *Sync {
	return &Sync{
		namerd:    n,
		store:     s,
		publisher: p,
	}
}

func syncDtab(ctx context.Context, s store.DentriesByDtabSelector, nd namerd.DentriesUpdator, dtab namerd.Dtab) error {
	dC, err := s.DentriesByDtab(ctx, dtab.String())
	if err != nil {
		return err
	}
	var dentries namerd.Dentries
	for dentry := range dC {
		dentries = append(dentries, namerd.Dentry{
			Prefix: dentry.Prefix,
			Dst:    dentry.Destination,
		})
	}
	return nd.UpdateDentries(dtab, dentries)
}
