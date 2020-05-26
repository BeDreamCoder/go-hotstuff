package transport

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/zhigui-projects/go-hotstuff/common/log"
	"github.com/zhigui-projects/go-hotstuff/pb"
	"google.golang.org/grpc/peer"
)

var logger = log.GetLogger("module", "transport")

type BroadcastServer interface {
	pb.AtomicBroadcastServer
	BroadcastMsg(msg *pb.Message) error
	UnicastMsg(msg *pb.Message, dest int64) error
}

type abServer struct {
	sendChan map[int64]chan<- *pb.Message
	sendLock *sync.RWMutex
}

func NewABServer() BroadcastServer {
	return &abServer{
		sendChan: make(map[int64]chan<- *pb.Message),
		sendLock: new(sync.RWMutex),
	}
}

func (a *abServer) Broadcast(srv pb.AtomicBroadcast_BroadcastServer) error {
	addr, src := extractRemoteAddress(srv.Context())
	logger.Debug("Starting new broadcast handler for remote peer", "addr", addr, "replicaId", src)

	ch := make(chan *pb.Message)
	a.sendLock.Lock()
	if oldChan, ok := a.sendChan[src]; ok {
		logger.Debug("create new connection from replica node", "replicaId", src)
		close(oldChan)
	}
	a.sendChan[src] = ch
	a.sendLock.Unlock()

	// TODO: firstly connect sync data

	var err error
	for msg := range ch {
		if err = srv.Send(msg); err != nil {
			a.sendLock.Lock()
			delete(a.sendChan, src)
			a.sendLock.Unlock()
			logger.Error("disconnected with replica node", "replicaId", src, "error", err)
		}
	}

	return err
}

func (a *abServer) BroadcastMsg(msg *pb.Message) error {
	a.sendLock.RLock()
	defer a.sendLock.Unlock()

	for _, ch := range a.sendChan {
		ch <- msg
	}
	return nil
}

func (a *abServer) UnicastMsg(msg *pb.Message, dest int64) error {
	a.sendLock.RLock()
	defer a.sendLock.Unlock()

	ch, ok := a.sendChan[dest]
	if !ok {
		logger.Error("unicast msg to invalid replica node", "replicaId", dest)
		return errors.Errorf("unicast msg to invalid replica node: %d", dest)
	}
	ch <- msg
	return nil
}

func extractRemoteAddress(ctx context.Context) (remoteAddress string, replicaId int64) {
	if value, ok := ctx.Value(nodeKey).(*NodeContext); ok {
		replicaId = int64(value.replicaId)
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", 0
	}
	if address := p.Addr; address != nil {
		remoteAddress = address.String()
	}
	return
}
