package link

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Link struct {
	URL string
	Key Key
}

type Key string

func (key Key) String() string { return strings.TrimSpace(string(key)) }

func (key Key) Empty() bool { return key.String() == "" }

func (link Link) Validate() error {
	if err := link.Key.Validate(); err != nil {
		return fmt.Errorf("key: %w", err)
	}
	u, err := url.ParseRequestURI(link.URL)
	if err != nil {
		return err
	}
	if u.Host == "" {
		return errors.New("empty host")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("scheme must be http or https")
	}
	return nil
}

func (key Key) Validate() error {
	const maxKeyLen = 16
	if keyLen := len(key.String()); keyLen > maxKeyLen {
		return fmt.Errorf("too long: len=%d, max=%d)", keyLen, maxKeyLen)
	}
	return nil
}

func Shorten(link Link) (Key, error) {
	if link.Key.Empty() {
		sum := sha256.Sum256([]byte(link.URL))
		link.Key = Key(base64.RawURLEncoding.EncodeToString(sum[:6]))
	}
	if err := link.Validate(); err != nil {
		return "", fmt.Errorf("validating: %w", err)
	}
	return link.Key, nil
}
