package main

import (
	"errors"
	"github.com/jackc/pgproto3/v2"
	"math/rand"
	"net"
)

type AuthMapper = func(props map[string]string, password string, salt [4]byte) (map[string]string, string, error)

var authError = errors.New("client_auth_failed")

var pgErrorAuthFailed = &pgproto3.ErrorResponse{
	Severity: "ERROR", Code: "28P01", Message: "password authentication failed",
}

type ProxyFront struct {
	salt        [4]byte
	conn        net.Conn
	proto       *pgproto3.Backend
	logger      *CustomLoggerHolder
	authMapper  AuthMapper
	originProps map[string]string
	mappedProps map[string]string
	mappedPass  string
}

func (f *ProxyFront) generateSalt() {
	r := make([]byte, 4)
	rand.Read(r)
	copy(f.salt[:], r)
}

func NewProxyFront(conn net.Conn, authMapper AuthMapper) *ProxyFront {
	return &ProxyFront{
		salt:       [4]byte{},
		conn:       conn,
		proto:      pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn),
		logger:     NewLoggerHolder(map[string]interface{}{"remote_address": conn.RemoteAddr().String()}),
		authMapper: authMapper,
	}
}

func (f *ProxyFront) handlePasswordAuth(msg *pgproto3.PasswordMessage) error {
	props, mappedPass, err := f.authMapper(f.originProps, msg.Password, f.salt)
	f.mappedProps = props
	f.mappedPass = mappedPass

	if err != nil {
		_ = f.proto.Send(pgErrorAuthFailed)
		return authError
	}
	err = f.proto.Send(&pgproto3.AuthenticationOk{})
	if err != nil {
		return err
	}
	return f.proto.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
}

func (f *ProxyFront) handleStartup() (map[string]string, error) {
	startupMessage, err := f.proto.ReceiveStartupMessage()
	if err != nil {
		f.logger.get().Warn("Failed to receive startup message")
		return nil, err
	}
	switch startupMessage.(type) {
	case *pgproto3.StartupMessage:
		msg := startupMessage.(*pgproto3.StartupMessage)
		f.originProps = msg.Parameters
		f.logger.setProperty("origin_database", f.originProps["database"])
		f.logger.setProperty("origin_user", f.originProps["user"])
		f.logger.setProperty("client_app", f.originProps["application_name"])
		f.generateSalt()
		err := f.proto.Send(&pgproto3.AuthenticationMD5Password{Salt: f.salt})
		if err != nil {
			f.logger.get().Error("Failed to send md5 password authentication request")
			return nil, err
		}
		return f.originProps, nil
	case *pgproto3.SSLRequest:
		_, err = f.conn.Write([]byte("N"))
		if err != nil {
			f.logger.get().Error("Failed to send deny SSL")
			return nil, err
		}
		return f.handleStartup()
	default:
		f.logger.get().Error("Failed to decode startup message")
		return nil, net.UnknownNetworkError("Failed to decode startup message")
	}
}

func (f *ProxyFront) Run() {
	f.logger.get().Debug("Client connected")
	defer f.logger.get().Debug("Connection closed")
	_, err := f.handleStartup()
	if err != nil {
		f.logger.get().WithError(err).Warn("Client connection init error")
	}
	f.logger.get().Info("Client startup sequence complete")
	for {
		msg, err := f.proto.Receive()
		if err != nil {
			f.logger.get().Warn("Error reading message")
			return
		}
		switch msg.(type) {
		case *pgproto3.PasswordMessage:
			err := f.handlePasswordAuth(msg.(*pgproto3.PasswordMessage))
			if err != nil {
				f.logger.get().WithError(err).Error("Failed to authenticate")
				return
			}
			f.logger.get().Debug("Authentication successful")
			//break
		default:
			f.logger.get().Error("Unknown message type")
			return
		}
	}
}
