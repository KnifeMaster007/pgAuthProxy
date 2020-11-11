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

func saltedMd5Credential(user string, password string, salt [4]byte) string {
	credHash := createMd5Credential(user, password)[3:]
	saltedCredHash := md5.Sum(append([]byte(credHash), salt[:]...))
	return "md5" + hex.EncodeToString(saltedCredHash[:])
}

func authStub(props map[string]string, password string, salt [4]byte) (map[string]string, string, error) {
	const username = "testuser"
	const pass = "password"
	const mappedUser = "postgres"
	const mappedPassword = "postgres"
	const mappedDatabase = "postgres"
	if props["user"] == username {
		if password != saltedMd5Credential(username, pass, salt) {
			return nil, "", io.EOF
		}

		mappedProps := make(map[string]string)
		for k, v := range props {
			mappedProps[k] = v
		}
		mappedProps["user"] = mappedUser
		mappedProps["database"] = mappedDatabase
		return mappedProps, mappedPassword, nil
	}
	return nil, mappedPassword, io.EOF
}
