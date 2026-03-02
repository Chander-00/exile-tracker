package ssh

import (
	"github.com/charmbracelet/ssh"
	"github.com/ByChanderZap/exile-tracker/config"
)

// PublicKeyHandler allows all SSH connections regardless of key.
func PublicKeyHandler(_ ssh.Context, _ ssh.PublicKey) bool {
	return true
}

// IsAdmin checks whether the session's public key matches the configured admin key.
func IsAdmin(sess ssh.Session) bool {
	adminKeyStr := config.Envs.SSHAdminKey
	if adminKeyStr == "" {
		return false
	}

	sessionKey := sess.PublicKey()
	if sessionKey == nil {
		return false
	}

	allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(adminKeyStr))
	if err != nil {
		return false
	}

	return ssh.KeysEqual(sessionKey, allowed)
}
