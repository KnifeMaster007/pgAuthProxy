package proxy

import (
	"encoding/binary"
	"errors"
	"github.com/jackc/pgproto3/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"pgAuthProxy/auth"
	"pgAuthProxy/utils"
	"strings"
	"time"
)

type ProxyBack struct {
	TargetProps map[string]string
	TargetHost  string

	originProps      map[string]string
	backendConn      net.Conn
	proto            *pgproto3.Frontend
	protoChunkReader pgproto3.ChunkReader
}

const MaxTcpPayload = 65535

var (
	MissingRequiredTargetFields = errors.New("required target fields missing in target props")
	BackendAuthenticationError  = errors.New("backend authentication failed")
	BackendInvalidMessage       = errors.New("unexpected message received from backend")
)

func NewProxyBackend(targetProps map[string]string, originProps map[string]string) (*ProxyBack, error) {
	b := &ProxyBack{
		originProps: originProps,
		TargetProps: make(map[string]string),
	}
	if host, ok := targetProps[utils.TargetHostParameter]; ok {
		b.TargetHost = host
	} else {
		return nil, MissingRequiredTargetFields
	}
	for k, v := range targetProps {
		if !strings.HasPrefix(k, utils.MetaPrefix) {
			b.TargetProps[k] = v
		}
	}
	err := b.initiateBackendConnection(targetProps[utils.TargetCredentialParameter])
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (b *ProxyBack) initiateBackendConnection(credential string) error {
	conn, err := net.Dial("tcp", b.TargetHost)
	if err != nil {
		return err
	}
	b.backendConn = conn
	b.protoChunkReader = pgproto3.NewChunkReader(conn)
	b.proto = pgproto3.NewFrontend(b.protoChunkReader, conn)
	err = b.proto.Send(&pgproto3.StartupMessage{
		ProtocolVersion: pgproto3.ProtocolVersionNumber,
		Parameters:      b.TargetProps,
	})
	if err != nil {
		conn.Close()
		return err
	}
	for {
		msg, err := b.proto.Receive()
		if err != nil {
			conn.Close()
			return err
		}
		switch msg.(type) {
		case *pgproto3.AuthenticationMD5Password:
			salt := msg.(*pgproto3.AuthenticationMD5Password).Salt
			err = b.proto.Send(&pgproto3.PasswordMessage{Password: auth.SaltedMd5Credential(credential, salt)})
			if err != nil {
				conn.Close()
				return err
			}
			continue
		case *pgproto3.AuthenticationOk:
			return nil
		case *pgproto3.ErrorResponse:
			return BackendAuthenticationError
		default:
			conn.Close()
			return BackendInvalidMessage
		}
	}
}

func pipeBackendPgMessages(source pgproto3.ChunkReader, dest io.Writer) error {
	bw := utils.NewBufferedWriter(MaxTcpPayload, dest)
	pipeRunning := true
	var pipeError error
	defer func() { pipeRunning = false }()

	go func() {
		for pipeRunning {
			time.Sleep(100 * time.Millisecond)
			_, err := bw.Flush()
			if err != nil {
				pipeRunning = false
				pipeError = err
			}
		}
	}()

	for {
		if pipeError != nil {
			return pipeError
		}
		header, err := source.Next(5)
		if err != nil {
			return err
		}
		l := int(binary.BigEndian.Uint32(header[1:])) - 4
		body, _ := source.Next(l)
		_, err = bw.Write(append(header, body...))
		if err != nil {
			return err
		}
	}
}

func pipePgMessages(source pgproto3.ChunkReader, dest io.Writer) error {
	for {
		header, err := source.Next(5)
		if err != nil {
			return err
		}
		l := int(binary.BigEndian.Uint32(header[1:])) - 4
		body, _ := source.Next(l)
		_, err = dest.Write(append(header, body...))
		if err != nil {
			return err
		}
	}
}

func (b *ProxyBack) Run(frontConn net.Conn, frontChunkReader pgproto3.ChunkReader) error {
	defer b.Close()
	err := make(chan error)
	go func() {
		log.Debug("bootstrapped backend -> frontend message pipe")
		err <- pipeBackendPgMessages(b.protoChunkReader, frontConn)
	}()
	go func() {
		log.Debug("bootstrapped backend <- frontend message pipe")
		err <- pipePgMessages(frontChunkReader, b.backendConn)
	}()
	select {
	case e := <-err:
		return e
	}
}

func (b *ProxyBack) Close() {
	if b.backendConn != nil {
		b.backendConn.Close()
	}
}
