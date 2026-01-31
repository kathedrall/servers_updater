package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"servers_updater/internal/domain"
	"time"

	"go.etcd.io/bbolt"
)

type BoltDB struct {
	conn *bbolt.DB
}

const (
	BUCKET_SETTINGS    = "Settings"
	BUCKET_CREDENTIALS = "Credentials"
	BUCKET_RECEPIENTS  = "Recipients"
	BUCKET_NAME        = "Config"
	BUCKET_MACHINES    = "Machines"
	SMTP_KEY           = "smtp_config"
	POOL_LIMIT_KEY     = "pool_limit"
	POOL_LIMIT_DEFAULT = 10

	SSH_USER_KEY     = "ssh_user"
	SSH_KEY_PATH_KEY = "ssh_key_path"
)

func InitDB(path string) (*BoltDB, error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		e := fmt.Errorf("error when starting bbolt: %w", err)
		return nil, e
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{
			BUCKET_SETTINGS,
			BUCKET_CREDENTIALS,
			BUCKET_MACHINES,
			BUCKET_RECEPIENTS,
			BUCKET_NAME,
		}

		for _, bucketName := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &BoltDB{conn: db}, nil
}

func (db *BoltDB) GetAllMachines() ([]domain.Machine, error) {
	var machines []domain.Machine

	err := db.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_MACHINES))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var m domain.Machine
			if err := json.Unmarshal(v, &m); err != nil {
				e := fmt.Errorf("Error reading the machines: %v", err)
				return e
			}
			machines = append(machines, m)
			return nil
		})
	})

	return machines, err
}

func (db *BoltDB) SaveMachine(m domain.Machine) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_MACHINES))
		if err != nil {
			return err
		}

		key := m.Host
		if m.Id != "" {
			key = m.Id
		}

		data, err := json.Marshal(m)
		if err != nil {
			e := fmt.Errorf("FAILED to convert to json: %v", err)
			return e
		}

		return bucket.Put([]byte(key), data)
	})
}

func (db *BoltDB) Close() error {
	return db.conn.Close()
}

func (db *BoltDB) SaveSMTPConfig(config domain.SMTPConfig) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_NAME))
		if err != nil {
			return err
		}
		data, _ := json.Marshal(config)
		return bucket.Put([]byte(SMTP_KEY), data)
	})
}

func (db *BoltDB) GetSMTPConfig() (domain.SMTPConfig, error) {
	var config domain.SMTPConfig
	err := db.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		if bucket == nil {
			return fmt.Errorf("Config Bucket not found.")
		}

		data := bucket.Get([]byte("smtp_config"))
		if data == nil {
			return fmt.Errorf("no SMTP config found")
		}
		return json.Unmarshal(data, &config)
	})
	return config, err
}

func (db *BoltDB) AddRecipient(email string) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_RECEPIENTS))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(email), []byte(email))
	})
}

func (db *BoltDB) RemoveRecipient(email string) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_RECEPIENTS))
		if bucket == nil {
			return nil
		}
		return bucket.Delete([]byte(email))
	})
}

func (db *BoltDB) ListRecipient() ([]string, error) {
	var emails []string
	err := db.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_RECEPIENTS))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			emails = append(emails, string(v))
			return nil
		})
	})
	return emails, err
}

func (db *BoltDB) SavePassword(host string, password string, masterKey []byte) error {
	encrypted, err := encrypt([]byte(password), masterKey)
	if err != nil {
		return err
	}
	return db.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_CREDENTIALS))
		return b.Put([]byte(host), encrypted)
	})
}

func (db *BoltDB) GetPassword(host string, masterKey []byte) (string, error) {
	var decrypted []byte
	err := db.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_CREDENTIALS))
		v := b.Get([]byte(host))
		if v == nil {
			return fmt.Errorf("Password not found for host.: %s", host)
		}

		var errDec error
		decrypted, errDec = decrypt(v, masterKey)
		return errDec
	})
	return string(decrypted), err
}

func encrypt(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plainText, nil), nil
}

func decrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := cipherText[:nonceSize], cipherText[nonceSize:]
	return gcm.Open(nil, nonce, actualCiphertext, nil)
}

func (db *BoltDB) SavePoolLimit(limit int) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_SETTINGS))
		val := fmt.Sprintf("%d", limit)
		return b.Put([]byte(POOL_LIMIT_KEY), []byte(val))
	})
}

func (db *BoltDB) GetPoolLimit() (int, error) {
	var limit int
	err := db.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_SETTINGS))
		v := b.Get([]byte(POOL_LIMIT_KEY))
		if v == nil {
			limit = POOL_LIMIT_DEFAULT
			return nil
		}
		_, err := fmt.Sscanf(string(v), "%d", &limit)
		return err
	})
	return limit, err
}

func (db *BoltDB) SaveSSHConfig(user string, keyPath string) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_SETTINGS))
		if b == nil {
			return fmt.Errorf("bucket %s not found", BUCKET_SETTINGS)
		}

		if err := b.Put([]byte(SSH_USER_KEY), []byte(user)); err != nil {
			return err
		}
		return b.Put([]byte(SSH_KEY_PATH_KEY), []byte(keyPath))
	})
}

func (db *BoltDB) GetSSHConfig() (string, string, error) {
	var user, keyPath string

	err := db.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_SETTINGS))
		if b == nil {
			return fmt.Errorf("bucket %s not found", BUCKET_SETTINGS)
		}

		vUser := b.Get([]byte(SSH_USER_KEY))
		vKey := b.Get([]byte(SSH_KEY_PATH_KEY))
		if vUser == nil || vKey == nil {
			return errors.New("SSH configuration not found.")
		}

		user = string(vUser)
		keyPath = string(vKey)
		return nil
	})
	return user, keyPath, err
}
