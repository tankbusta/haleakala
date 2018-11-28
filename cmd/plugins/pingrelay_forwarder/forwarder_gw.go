package pingrelay_forwarder

import (
	"log"
	"sync"
	"syscall"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
)

const moarMessages = zmq.Errno(syscall.EAGAIN)

type Context struct {
	cfg   Config
	stop  chan bool
	pulls *zmq.Socket
	pushs *zmq.Socket
	wg    *sync.WaitGroup
}

func New(cfg Config) (*Context, error) {
	zmq.AuthSetVerbose(true)
	if err := zmq.AuthStart(); err != nil {
		return nil, err
	}

	pulls, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}

	if err := pulls.SetSubscribe(""); err != nil {
		return nil, err
	}

	pushs, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}

	if err := pushs.ServerAuthCurve(cfg.ServerDomain, cfg.ServerKey); err != nil {
		return nil, err
	}

	return &Context{
		stop:  make(chan bool, 1),
		cfg:   cfg,
		wg:    &sync.WaitGroup{},
		pulls: pulls,
		pushs: pushs,
	}, nil
}

func (s *Context) mainloop() {
	defer s.wg.Done()

	var b []byte
	var err error

	s.pulls.SetRcvtimeo(1 * time.Second)

Loop:
	for {
		select {
		case _ = <-s.stop:
			break Loop
		default:
		}

		b, err = s.pulls.RecvBytes(0)
		if err != nil {
			// EAGAIN/moarMessages means the recv timeout has been reached
			if err == moarMessages {
				continue
			}
			log.Printf("An error occured while reading from PULL socket: %s\n", err)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"msg": string(b),
		}).Debug("Forwarder sending message")

		if _, err := s.pushs.SendBytes(b, 0); err != nil {
			log.Printf("An error occured while sending data to PUSH socket: %s\n", err)
		}
	}
}

// Start the security gateway
func (s *Context) Start() error {
	for _, cert := range s.cfg.AuthorizedKeys {
		log.Printf("Adding domain %s with addresses %v\n", s.cfg.ServerDomain, cert.Addressess)
		zmq.AuthAllow(s.cfg.ServerDomain, cert.Addressess...)
		zmq.AuthCurveAdd(s.cfg.ServerDomain, cert.PublicKey)
	}

	if err := s.pulls.Connect(s.cfg.ClientAddress); err != nil {
		return err
	}

	if err := s.pushs.Bind(s.cfg.ListenAddress); err != nil {
		return err
	}

	// start the relay loop!
	s.wg.Add(1)
	go s.mainloop()

	return nil
}

// Stop the security gateway
func (s *Context) Stop() error {
	close(s.stop)
	s.wg.Wait()

	zmq.AuthStop()

	if err := s.pulls.Close(); err != nil {
		log.Printf("Failed to close PULL socket: %s\n", err)
	}

	if err := s.pushs.Close(); err != nil {
		log.Printf("Failed to close PUSH socket: %s\n", err)
	}
	return nil
}
