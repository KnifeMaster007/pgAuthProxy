package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func createMd5Credential(user string, password string) string {
	credHash := md5.Sum([]byte(password + user))
	return "md5" + hex.EncodeToString(credHash[:])
}

func saltedMd5Credential(cred string, salt [4]byte) string {
	saltedCredHash := md5.Sum(append([]byte(cred[3:]), salt[:]...))
	return "md5" + hex.EncodeToString(saltedCredHash[:])
}

func saltedMd5PasswordCredential(user string, password string, salt [4]byte) string {
	return saltedMd5Credential(createMd5Credential(user, password), salt)
}

func authStub(props map[string]string, password string, salt [4]byte) (map[string]string, error) {
	const username = "testuser"
	const pass = "password"
	const mappedUser = "igalkin"
	const mappedPassword = "SnwUD5pS9Z4N"
	const mappedDatabase = "m4"
	const targetHost = "pgbouncer01.d.m4"
	const targetPort = "5432"
	var mappedCred = createMd5Credential(mappedUser, mappedPassword)
	if props["user"] == username {
		if password != saltedMd5PasswordCredential(username, pass, salt) {
			return nil, io.EOF
		}

		mappedProps := make(map[string]string)
		for k, v := range props {
			mappedProps[k] = v
		}
		mappedProps["user"] = mappedUser
		mappedProps["database"] = mappedDatabase
		mappedProps[TargetHostParameter] = targetHost
		mappedProps[TargetPortParameter] = targetPort
		mappedProps[TargetCredentialParameter] = mappedCred
		return mappedProps, nil
	}
	return nil, io.EOF
}
