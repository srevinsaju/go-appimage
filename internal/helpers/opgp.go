// TODO: Discuss with AppImage team whether we can switch from GPG to RSA
// and whether this would simplify things and reduce dependencies
// https://socketloop.com/tutorials/golang-saving-private-and-public-key-to-files

package helpers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	//	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func CreateAndValidateKeyPair() {
	createKeyPair()

	// error reading armored key openpgp:
	// invalid data: entity without any identities
	// b, err := ioutil.ReadFile("privkey")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// hexstring, _ := readPGP(b)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = validate(hexstring)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func createKeyPair() {
	conf := &packet.Config{
		RSABits: 4096,
	}
	entity, err := openpgp.NewEntity("Test Key", "Autogenerated GPG Key", "test@example.com", conf)
	if err != nil {
		log.Fatalf("error in entity.PrivateKey.Serialize(serializedEntity): %s", err)
	}
	// Generate private key and write it to a file
	serializedEntity := bytes.NewBuffer(nil)
	err = entity.PrivateKey.Serialize(serializedEntity)

	buf := bytes.NewBuffer(nil)
	headers := map[string]string{"Version": "GnuPG v1"}
	w, err := armor.Encode(buf, openpgp.PrivateKeyType, headers)
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(serializedEntity.Bytes())
	if err != nil {
		log.Fatalf("error armoring serializedEntity: %s", err)
	}
	w.Close()

	prf, err := os.Create(PrivkeyFileName)
	PrintError("ogpg", err)
	defer prf.Close()
	n2, err := prf.Write(buf.Bytes())
	PrintError("ogpg", err)

	fmt.Printf("wrote %d bytes\n", n2)

	// Generate public key and write it to a file
	serializedEntity = bytes.NewBuffer(nil)
	err = entity.PrimaryKey.Serialize(serializedEntity)
	if err != nil {
		log.Fatal(err)
	}
	buf = bytes.NewBuffer(nil)
	headers = map[string]string{"Version": "GnuPG v1"}
	w, err = armor.Encode(buf, openpgp.PublicKeyType, headers)
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(serializedEntity.Bytes())
	if err != nil {
		log.Fatalf("error armoring serializedEntity: %s", err)
	}
	w.Close()

	puf, err := os.Create(PubkeyFileName)
	PrintError("ogpg", err)
	defer puf.Close()
	n2, err = puf.Write(buf.Bytes())
	PrintError("ogpg", err)
	fmt.Printf("wrote %d bytes\n", n2)

}

func readPGP(armoredKey []byte) (string, error) {
	keyReader := bytes.NewReader(armoredKey)
	entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
	if err != nil {
		log.Fatalf("error reading armored key %s", err)
	}
	serializedEntity := bytes.NewBuffer(nil)
	err = entityList[0].Serialize(serializedEntity)
	if err != nil {
		return "", fmt.Errorf("error serializing entity for file %s", err)
	}

	return base64.StdEncoding.EncodeToString(serializedEntity.Bytes()), nil
}

// CheckSignature checks the signature embedded in an AppImage at path,
// returns the entity that has signed the AppImage and error
// based on https://stackoverflow.com/a/34008326
func CheckSignature(path string) (*openpgp.Entity, error) {
	var ent *openpgp.Entity
	err := errors.New("could not verify AppImage signature") // Be pessimistic by default, unless we can positively verify the signature
	pubkeybytes, err := GetSectionData(path, ".sig_key")

	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pubkeybytes))
	if err != nil {
		return ent, err
	}

	sigbytes, err := GetSectionData(path, ".sha256_sig")

	ent, err = openpgp.CheckArmoredDetachedSignature(keyring, strings.NewReader(CalculateSHA256Digest(path)), bytes.NewReader(sigbytes))
	if err != nil {
		return ent, err
	}

	return ent, nil
}
