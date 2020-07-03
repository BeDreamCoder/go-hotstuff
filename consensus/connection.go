package consensus

import (
	"io"
	"time"

	"github.com/zhigui-projects/go-hotstuff/pb"
	"github.com/zhigui-projects/go-hotstuff/transport"
)

type NodeInfo struct {
	Id        ReplicaID
	Addr      string
	TlsOpts   *transport.TLSOptions
	Connected bool
}

type NodeManager struct {
	transport.BroadcastServer
	*transport.GrpcServer
	Self *NodeInfo
	// all nodes info contains self
	Nodes map[ReplicaID]*NodeInfo
}

func NewNodeManager(id ReplicaID, replicas []*NodeInfo) *NodeManager {
	mgr := &NodeManager{
		Nodes: make(map[ReplicaID]*NodeInfo, len(replicas)),
	}
	for _, node := range replicas {
		if node.Id == id {
			grpcServer, err := transport.NewGrpcServer(node.Addr, node.TlsOpts)
			if err != nil {
				logger.Error("Failed to new grpc server", "error", err)
				panic(err)
			}

			server := transport.NewABServer()
			pb.RegisterAtomicBroadcastServer(grpcServer.Server(), server)
			mgr.BroadcastServer = server
			mgr.GrpcServer = grpcServer
			node.Connected = true
			mgr.Self = node
		}
		mgr.Nodes[node.Id] = node
	}

	return mgr
}

func (n *NodeManager) StartServer() {
	logger.Info("Hotstuff node started, beginning to serve requests", "replicaId", n.Self.Id, "serverAddress", n.Self.Addr)

	if err := n.Start(); err != nil {
		logger.Error(" Hotstuff node server start failed", "error", err)
		panic(err)
	}
}

func (n *NodeManager) ConnectWorkers(queue chan<- MsgExecutor) {
	for _, node := range n.Nodes {
		if node.Id == n.Self.Id {
			continue
		}
		go func(node *NodeInfo) {
			delay := time.After(0)
			for {
				node.Connected = false
				// reconnect attempts
				<-delay
				delay = time.After(transport.DefaultConnectionTimeout)

				logger.Debug("connecting to replica node", "id", node.Id, "address", node.Addr)

				bc, err := transport.NewBroadcastClient(node.Addr, int64(n.Self.Id), node.TlsOpts)
				if err != nil {
					logger.Warn("could not connect to replica node", "id", node.Id, "address", node.Addr, "error", err)
					continue
				}
				logger.Info("connection to replica node established", "id", node.Id, "address", node.Addr)
				node.Connected = true

				for {
					msg, err := bc.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						logger.Warn("consensus stream with replica node broke", "id", node.Id, "address", node.Addr, "error", err)
						break
					}

					go func() {
						queue <- &msgEvent{msg: msg, src: node.Id}
					}()
				}
			}
		}(node)
	}
}

func (n *NodeManager) GetConnectStatus(id int64) bool {
	node, ok := n.Nodes[ReplicaID(id)]
	if ok {
		return node.Connected
	} else {
		return false
	}
}
