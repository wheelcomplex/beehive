package bh

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
)

const (
	// TODO(soheil): Add all these as flags and configs.
	proxyProto                 = "tcp"
	minWaitInMs  time.Duration = 100
	maxWaitInMs                = 1000
	waitStepInMs               = 10
	maxRetries                 = 3
)

type proxyBee struct {
	localBee
	conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func (r *proxyBee) handleMsg(mh msgAndHandler) {
	mh.msg.MsgTo = r.rId
	if err := r.encoder.Encode(mh.msg); err != nil {
		glog.Errorf("Cannot encode message: %v", err)
	}
}

func dialHive(id HiveId) (net.Conn, error) {
	step := time.Duration(waitStepInMs)
	waitMs := minWaitInMs
	retries := 0
	for {
		c, err := net.Dial(proxyProto, string(id))
		if err == nil {
			return c, nil
		}

		if retries >= maxRetries {
			return nil, errors.New(
				fmt.Sprintf("Cannot connect to %+v: %+v. Failed %v times.", id, err,
					retries))
		}

		retries++
		// TODO(soheil): What if the pair has just crashed? We need to return error
		// and return from start and try to regrab the locks again. Then panic if
		// neither can be successful.
		glog.Errorf("Cannot connect to %+v: %+v. Retrying...", id, err)
		time.Sleep(waitMs * time.Millisecond)

		if waitMs > maxWaitInMs {
			waitMs = maxWaitInMs
			continue
		}

		if waitMs < maxWaitInMs {
			waitMs += step
			step *= 2
		}
	}

	return nil, errors.New(fmt.Sprintf("Cannot connect to remote hive: %+v", id))
}

func (r *proxyBee) dial() {
	// FIXME(soheil): This can't scale. We can only support 65k remote bees.
	// There should be one connection per remote hive and with that we can
	// support 65k remote controllers.
	conn, err := dialHive(r.rId.HiveId)
	if err != nil {
		glog.Fatalf("Cannot connect to peer: %v", err)
	}

	r.conn = conn
	r.encoder = gob.NewEncoder(r.conn)
	r.decoder = gob.NewDecoder(r.conn)

	if err = r.encoder.Encode(hiveHandshake{dataHandshake}); err != nil {
		glog.Fatalf("Cannot handshake with peer: %v", err)
	}

	if err = r.encoder.Encode(r.rId); err != nil {
		glog.Fatalf("Cannot encode bee: %v", err)
	}

	id := BeeId{}
	if err = r.decoder.Decode(&id); err != nil || r.rId != id {
		glog.Fatalf("Peer cannot find bee: %v", r.rId)
	}

	glog.V(2).Infof("Proxy connected to remote bee %+v", r.rId)
}

func (r *proxyBee) hangup() {
	if r.conn != nil {
		r.conn.Close()
	}
}

// TODO(soheil): Maybe start should return an error.
func (r *proxyBee) start() {
	r.dial()
	defer r.hangup()

	for {
		select {
		case d, ok := <-r.dataCh:
			if !ok {
				return
			}
			r.handleMsg(d)

		case c, ok := <-r.ctrlCh:
			if !ok {
				return
			}
			if ok = r.handleCmd(c); !ok {
				return
			}
		}
	}
}