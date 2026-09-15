// PasswordHash
package utils

import "golang.org/x/crypto/bcrypt"

type BycryptGroup struct {}

func (b *BycryptGroup) GeneratePasswordHash(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes)
}

func (b *BycryptGroup) CheckPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}