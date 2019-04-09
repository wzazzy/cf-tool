package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"syscall"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/client"
	"github.com/xalanq/cf-tool/util"
	"golang.org/x/crypto/ssh/terminal"
)

func createHash(key string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func encrypt(username, password string) (ret string, err error) {
	block, err := aes.NewCipher(createHash("glhf" + username + "233"))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	text := gcm.Seal(nonce, nonce, []byte(password), nil)
	ret = hex.EncodeToString(text)
	return
}

func decrypt(username, password string) (ret string, err error) {
	data, err := hex.DecodeString(password)
	if err != nil {
		err = errors.New("Cannot decode password")
		return
	}
	block, err := aes.NewCipher(createHash("glhf" + username + "233"))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonceSize := gcm.NonceSize()
	nonce, text := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, text, nil)
	if err != nil {
		return
	}
	ret = string(plain)
	return
}

// DecryptPassword get real password
func (c *Config) DecryptPassword() (string, error) {
	return decrypt(c.Username, c.Password)
}

// Login configurate
func (c *Config) Login(path string) (err error) {
	color.Cyan("Config username/email and password(encrypt)")

	fmt.Printf("username: ")
	username := util.ScanlineTrim()

	fmt.Printf("password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	password := string(bytePassword)
	fmt.Println()

	err = client.New(path).Login(username, password)
	if err != nil {
		return errors.New("Invalid username and password")
	}
	password, err = encrypt(username, password)
	if err == nil {
		c.Username = username
		c.Password = password
		err = c.save()
	}
	return
}
