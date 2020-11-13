package proxy

import (
	"errors"
	"fmt"
	"github.com/jackc/pgproto3/v2"
	"math/rand"
	"net"
	"pgAuthProxy/utils"
)

type AuthMapper = func(props map[string]string, password string, salt [4]byte) (map[string]string, error)

var authError = errors.New("client_auth_failed")

var pgErrorAuthFailed = &pgproto3.ErrorResponse{
	Severity: "ERROR", Code: "28P01", Message: "password authentication failed",
}

type ProxyFront struct {
	salt             [4]byte
	conn             net.Conn
	proto            *pgproto3.Backend
	protoChunkReader pgproto3.ChunkReader
	logger           *utils.CustomLoggerHolder
	authMapper       AuthMapper
	originProps      map[string]string
	mappedProps      map[string]string
	backend          *ProxyBack
}

func (f *ProxyFront) generateSalt() {
	r := make([]byte, 4)
	rand.Read(r)
	copy(f.salt[:], r)
}

func NewProxyFront(conn net.Conn, authMapper AuthMapper) *ProxyFront {
	cr := pgproto3.NewChunkReader(conn)
	return &ProxyFront{
		salt:             [4]byte{},
		conn:             conn,
		proto:            pgproto3.NewBackend(cr, conn),
		protoChunkReader: cr,
		logger:           utils.NewLoggerHolder(map[string]interface{}{"remote_address": conn.RemoteAddr().String()}),
		authMapper:       authMapper,
	}
}

func (f *ProxyFront) handlePasswordAuth(msg *pgproto3.PasswordMessage) error {
	props, err := f.authMapper(f.originProps, msg.Password, f.salt)
	if err != nil {
		_ = f.proto.Send(pgErrorAuthFailed)
		return authError
	}
	f.mappedProps = props
	f.backend, err = NewProxyBackend(f.mappedProps, f.originProps)
	if err != nil {
		f.logger.Get().Error("Failed to bootstrap backend connection")
	}
	f.logger.SetProperty("targetDsn", fmt.Sprintf(
		"postgres://%s@%s:%s/%s",
		f.backend.TargetProps["user"],
		f.backend.TargetHost,
		f.backend.TargetPort,
		f.backend.TargetProps["database"],
	))
	f.logger.Get().Debug("Bootstrapped backend connection")
	err = f.proto.Send(&pgproto3.AuthenticationOk{})
	if err != nil {
		return err
	}
	return nil
}

func (f *ProxyFront) handleStartup() (map[string]string, error) {
	startupMessage, err := f.proto.ReceiveStartupMessage()
	if err != nil {
		f.logger.Get().Warn("Failed to receive startup message")
		return nil, err
	}
	switch startupMessage.(type) {
	case *pgproto3.StartupMessage:
		msg := startupMessage.(*pgproto3.StartupMessage)
		f.originProps = msg.Parameters
		f.logger.SetProperty("origin_database", f.originProps["database"])
		f.logger.SetProperty("origin_user", f.originProps["user"])
		f.logger.SetProperty("client_app", f.originProps["application_name"])
		f.generateSalt()
		err := f.proto.Send(&pgproto3.AuthenticationMD5Password{Salt: f.salt})
		if err != nil {
			f.logger.Get().Error("Failed to send md5 password authentication request")
			return nil, err
		}
		return f.originProps, nil
	case *pgproto3.SSLRequest:
		_, err = f.conn.Write([]byte("N"))
		if err != nil {
			f.logger.Get().Error("Failed to send deny SSL")
			return nil, err
		}
		return f.handleStartup()
	default:
		f.logger.Get().Error("Failed to decode startup message")
		return nil, net.UnknownNetworkError("Failed to decode startup message")
	}
}

func (f ProxyFront) Close() {
	if f.backend != nil {
		f.backend.Close()
	}
}

func (f *ProxyFront) Run() {
	f.logger.Get().Debug("Client connected")
	defer f.logger.Get().Debug("Connection closed")
	_, err := f.handleStartup()
	if err != nil {
		f.logger.Get().WithError(err).Warn("Client connection init error")
	}
	f.logger.Get().Info("Client startup sequence complete")
	for {
		msg, err := f.proto.Receive()
		if err != nil {
			f.logger.Get().Warn("Error reading message")
			return
		}
		switch msg.(type) {
		case *pgproto3.PasswordMessage:
			err := f.handlePasswordAuth(msg.(*pgproto3.PasswordMessage))
			if err != nil {
				f.logger.Get().WithError(err).Error("Failed to authenticate")
				return
			}
			f.logger.Get().Debug("Authentication successful")
			goto serve
		default:
			f.logger.Get().Error("Unknown message type")
			return
		}
	}
serve:
	err = f.backend.Run(f.conn, f.protoChunkReader)
}
