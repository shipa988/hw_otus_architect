package queue

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/usecases"
	"os"
	"time"
)

const (
	connectNATSErr      = `can't connect to nats server %v`
	connectSTANErr      = `can't connect to NATS Streaming Server at: %s`
	pushErr             = `can't async publish to nats`
	pullErr             = `can't subscribe on messages from nats`
	pushNonEmptyGuidErr = `expected non-empty guid to be returned`
	ackPushErr          = `error in server ack for guid %s: %v`
	ackMsgErr           = `failed to ACK msg: %d`
	ackPushGuidErr      = `expected a matching guid in ack callback, got %s vs %s`
	processingErr       = `processing message from stan error`
	doneProcessing      = "STAN stopping processing messages by context done"
)

var _ usecases.Queue = (*STANQueue)(nil)


type STANQueue struct {
	nc      *nats.Conn
	sc      stan.Conn
	stancfg StanConnection
	natscfg NatsConnection
}

func NewSTANQueue(stancfg StanConnection, natscfg NatsConnection) *STANQueue {
	return &STANQueue{
		stancfg: stancfg,
		natscfg: natscfg,
	}
}

func (q *STANQueue) Connect(ctx context.Context) error {
	var err error
	q.nc, err = nats.Connect(
		q.natscfg.Address+`:`+q.natscfg.Port,
		nats.Name(q.natscfg.Name),
		nats.Timeout(time.Millisecond*q.natscfg.TimeoutMS),
		nats.PingInterval(time.Millisecond*q.natscfg.PingIntervalMS),
		nats.MaxPingsOutstanding(q.natscfg.MaxPingsOutstanding),
		nats.ReconnectWait(time.Millisecond*q.natscfg.ReconnectWait),
		nats.ReconnectBufSize(q.natscfg.ReconnectBufSize),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Error(errors.Wrap(err, "Disconnect Handler"))
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("Got reconnected to %v!\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("Connection closed. Reason: %q\n", nc.LastError())
		}),
	)
	if err != nil {
		return errors.Wrapf(err, connectNATSErr, q.natscfg.Name)
	}
	log.Info("Connected to NATS %v", q.natscfg.Address+`:`+q.natscfg.Port)

	// Streaming Connection
	q.sc, err = stan.Connect(
		q.stancfg.ClusterID,
		q.stancfg.ClientID,
		stan.NatsConn(q.nc),
		stan.MaxPubAcksInflight(q.stancfg.MaxPubAcksInFlight),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Info("Connection lost, reason: %v\n", reason)
			log.Info("Draining...")
			q.nc.Drain()
			log.Info("Exiting")
			os.Exit(0)
		}),
	)
	if err != nil {
		return errors.Wrapf(err, connectSTANErr, q.natscfg.Address+`:`+q.natscfg.Port)
	}
	log.Info("Connected to NATS-Streaming %v", q.stancfg.ClusterID+`:`+q.stancfg.ClientID)

	return nil
}

func (q *STANQueue) Disconnect(ctx context.Context) {
	q.nc.Close()
	q.sc.Close()
}

func (q *STANQueue) Push(ctx context.Context, msg []byte) error {
	var guid string
	guid, err := q.sc.PublishAsync(q.stancfg.SubjectPublish, msg, func(lguid string, err error) {
		if err != nil {
			log.Error(errors.Wrapf(err, ackPushErr, lguid))
		}
		if lguid != guid {
			log.Error(ctx, ackPushGuidErr, lguid, guid)
		}
	})
	if err != nil {
		return errors.Wrap(err, pushErr)
	}

	if guid == "" {
		return errors.New(pushNonEmptyGuidErr)
	}
	return nil
}

func (q *STANQueue) Pull(ctx context.Context, timeToFlush time.Duration, batchSize int, clb func(ctx context.Context, msgs [][]byte) error) error {
	messagesForAcks := make([]*stan.Msg, 0, batchSize)
	msgs := make(chan *stan.Msg)
	ticker := time.NewTicker(time.Second)
	flushTime := time.Now()

	buffer := make([][]byte, 0, batchSize)
	flush := func() {
		if len(messagesForAcks)>0 && (time.Now().Sub(flushTime) >= timeToFlush || len(messagesForAcks) >= batchSize) {
			flushTime = time.Now()
			err := clb(ctx, buffer)
			if err != nil {
				log.Error(errors.Wrap(err, processingErr))
				return
			}
			for i, _ := range messagesForAcks {
				if err := messagesForAcks[i].Ack(); err != nil {
					log.Error(errors.Wrapf(err, ackMsgErr, messagesForAcks[i].Sequence))
					return
				}
			}
			buffer = buffer[:0]
			messagesForAcks = messagesForAcks[:0]
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info(doneProcessing)
				return
			default:
			}
			select {
			case <-ctx.Done():
				log.Info(doneProcessing)
				return
			case <-ticker.C:
				flush()
			case msg, ok := <-msgs:
				if !ok {
					log.Info(doneProcessing)
					return
				}
				messagesForAcks = append(messagesForAcks, msg)
				buffer = append(buffer, msg.Data)
				flush()
			}
		}
	}()
	_, err := q.sc.QueueSubscribe(q.stancfg.SubjectPublish, q.stancfg.GroupName, func(msg *stan.Msg) {
		select {
		//close connection if done
		case <-ctx.Done():
			log.Info("exit from pulling by context done")
			err := q.sc.Close()
			if err != nil {
				log.Error(errors.Wrapf(err, "error while exit from pulling by context done"))
			}
			return
		default:
		}
		msgs <- msg
	}, stan.DeliverAllAvailable(), stan.DurableName(q.stancfg.DurableName), stan.SetManualAckMode(),
		stan.MaxInflight(q.stancfg.MaxWithoutAck), stan.AckWait(time.Duration(q.stancfg.AckWaitTimeSec+q.stancfg.AckWaitDelay)*time.Second))
	if err != nil {
		e := q.sc.Close()
		if e != nil {
			log.Error(errors.Wrapf(e, pullErr))
		}
		return errors.Wrap(err, pullErr)
	}
	if err := q.nc.LastError(); err != nil {
		return errors.Wrap(err, pullErr)
	}

	return nil
}
