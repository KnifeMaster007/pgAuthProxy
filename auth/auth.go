package auth

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/viper"
	"io"
	"pgAuthProxy/utils"
	"strings"
)

func CreateMd5Credential(user string, password string) string {
	credHash := md5.Sum([]byte(password + user))
	return "md5" + hex.EncodeToString(credHash[:])
}

func SaltedMd5Credential(cred string, salt [4]byte) string {
	saltedCredHash := md5.Sum(append([]byte(cred[3:]), salt[:]...))
	return "md5" + hex.EncodeToString(saltedCredHash[:])
}

func SaltedMd5PasswordCredential(user string, password string, salt [4]byte) string {
	return SaltedMd5Credential(CreateMd5Credential(user, password), salt)
}

func encodeProps(props map[string]string) io.Reader {
	builder := &bytes.Buffer{}
	for k, v := range props {
		builder.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	return builder
}

func Exec(props map[string]string, password string, salt [4]byte) (map[string]string, error) {
	parameters := map[string]string{
		utils.SourceCredentialParameter: password,
		utils.SourceSaltParameter:       hex.EncodeToString(salt[:]),
	}
	for k, v := range props {
		parameters[k] = v
	}
	args := viper.GetStringSlice("authenticator.cmd")
	command := gocmd.NewCmd(args[0], args[1:]...)
	statusChan := command.StartWithStdin(encodeProps(parameters))

	select {
	case status := <-statusChan:
		if status.Error != nil {
			return nil, status.Error
		}
		if status.Exit != 0 {
			return nil, io.EOF
		}
		ret := make(map[string]string)
		for _, s := range status.Stdout {
			if split := strings.SplitN(s, "=", 2); len(split) == 2 {
				if !strings.HasPrefix(split[0], utils.SourceMetaPrefix) {
					ret[split[0]] = split[1]
				}
			}
		}
		return ret, nil
	}
}
