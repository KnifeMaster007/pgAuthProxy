package auth

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"pgAuthProxy/utils"
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

func AuthStub(props map[string]string, password string, salt [4]byte) (map[string]string, error) {
	const username = "testuser"
	const pass = "password"
	const mappedUser = "igalkin"
	const mappedPassword = "SnwUD5pS9Z4N"
	const mappedDatabase = "m4"
	const targetHost = "pgbouncer01.d.m4"
	const targetPort = "5432"
	var mappedCred = CreateMd5Credential(mappedUser, mappedPassword)
	if props["user"] == username {
		if password != SaltedMd5PasswordCredential(username, pass, salt) {
			return nil, io.EOF
		}

		mappedProps := make(map[string]string)
		for k, v := range props {
			mappedProps[k] = v
		}
		mappedProps["user"] = mappedUser
		mappedProps["database"] = mappedDatabase
		mappedProps[utils.TargetHostParameter] = targetHost
		mappedProps[utils.TargetPortParameter] = targetPort
		mappedProps[utils.TargetCredentialParameter] = mappedCred
		return mappedProps, nil
	}
	return nil, io.EOF
}
