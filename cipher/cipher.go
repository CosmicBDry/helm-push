package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

func Encrypt(plainText string, keyText string) ([]byte, error) {
	// 转换成字节数据, 方便加密
	plainByte := []byte(plainText)
	keyByte := []byte(keyText)
	// 创建加密算法aes
	c, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, err
	}
	//加密字符串
	cfb := cipher.NewCFBEncrypter(c, commonIV)
	cipherByte := make([]byte, len(plainByte))
	cfb.XORKeyStream(cipherByte, plainByte)

	return cipherByte, nil
}
func Decrypt(cipherByte []byte, keyText string) (string, error) {
	// 转换成字节数据, 方便加密
	keyByte := []byte(keyText)
	// 创建加密算法aes
	c, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	// 解密字符串
	cfbdec := cipher.NewCFBDecrypter(c, commonIV)
	plainByte := make([]byte, len(cipherByte))
	cfbdec.XORKeyStream(plainByte, cipherByte)
	plainText := string(plainByte)
	return plainText, nil
}

func HexDecrypt(hexPassword string) ([]byte, error) {
	cipherbyte, err := hex.DecodeString(hexPassword)
	if err != nil {
		return nil, err
	}
	return cipherbyte, nil

}
