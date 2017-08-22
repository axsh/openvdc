package vm

import (
	"encoding/json"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
)

type Base struct {
}

var SupportedAPICalls = []string{
	"/api.Instance/Start",
	"/api.Instance/Run",
	"/api.Instance/Stop",
	"/api.Instance/Reboot",
	"/api.Instance/Console",
	"/api.Instance/Log",
}

func (*Base) ValidateAuthenticationType(in json.RawMessage) (json.RawMessage, model.AuthenticationType, error) {
	var json_template struct {
		AuthenticationType string `json:"authentication_type,omitempty"`
	}

	if err := json.Unmarshal(in, &json_template); err != nil {
		return nil, 0, err
	}
	var ret model.AuthenticationType
	if json_template.AuthenticationType != "" {
		format, ok := model.AuthenticationType_value[strings.ToUpper(json_template.AuthenticationType)]
		if !ok {
			return nil, 0, errors.Errorf("Unknown value at format: %s", json_template.AuthenticationType)
		}
		ret = model.AuthenticationType(format)

		// Remove authentication_type field
		tmp := make(map[string]interface{})
		var err error
		if err = json.Unmarshal(in, &tmp); err != nil {
			return nil, 0, errors.Wrap(err, "Failed json.Unmarshal")
		}
		delete(tmp, "authentication_type")
		in, err = json.Marshal(tmp)
		if err != nil {
			return nil, 0, errors.Wrap(err, "Failed json.Marshal")
		}
	}
	return in, ret, nil
}

func (*Base) ValidatePublicKey(h handlers.ResourceHandler, authType model.AuthenticationType, sshPubKey string) error {
	switch authType {
	case model.AuthenticationType_NONE:
	case model.AuthenticationType_PUB_KEY:
		if sshPubKey == "" {
			return handlers.ErrInvalidTemplate(h, "ssh_public_key is not set")
		}

		err := validatePublicKey([]byte(sshPubKey))
		if err != nil {
			return handlers.ErrInvalidTemplate(h, err.Error())
		}

	default:
		return handlers.ErrInvalidTemplate(h, "Unknown authentication_type parameter"+authType.String())
	}
	return nil
}

func validatePublicKey(key []byte) error {
	keyStr := string(key[:])
	// Check that the key is in OpenSSH format.
	keyNames := []string{"ssh-rsa", "ssh-dss", "ecdsa-sha2-nistp256", "ssh-ed25519"}
	firstStr := strings.Fields(keyStr)
	for _, name := range keyNames {
		if firstStr[0] == name {
			return nil
		}
	}

	// // Check that the key is in SECSH format.
	// keyNames = []string{"SSH2 ", "RSA", ""}
	// for _, name := range keyNames {
	// 	if strings.Contains(keyStr, "---- BEGIN "+name+"PUBLIC KEY ----") &&
	// 		strings.Contains(keyStr, "---- END "+name+"PUBLIC KEY ----") {
	// 		return nil
	// 	}
	// }

	// Check that the key is in RFC4253 binary format.
	_, err := ssh.ParsePublicKey(key)
	if err != nil {
		return err
	}
	return nil
}

func (*Base) IsSupportAPI(method string) bool {
	for _, m := range SupportedAPICalls {
		if m == method {
			return true
		}
	}
	return false
}
