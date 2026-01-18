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
 "encoding/json"
 "go.etcd.io/bbolt"
)

type BoltDB struct {
 conn *bbolt.DB
}

const (
 BUCKET_SETTINGS = "Settings"
 BUCKET_CREDENTIALS = "Credentials"
 BUCKET_RECEPIENTS = "Recipients"
 BUCKET_NAME = "Config"
 SMTP_KEY = "smtp_config"
 POOL_LIMIT_KEY = "pool_limit"
 POOL_LIMIT_DEFAULT = 10

 SSH_USER_KEY = "ssh_user"
 SSH_KEY_PATH_KEY = "ssh_key_path"
)

func InitDB(path string) (*BoltDB, error) {
 db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second}) 
 if err != nil {
  return nil, fmt.Errorf("error when starting bbolt: %w", err)
 }

 err = db.Update(func(tx *bbolt.Tx) error {
  if _, err := tx.CreateBucketIfNotExists([]byte(BUCKET_SETTINGS)); err != nil {
   return err
 }
  if _, err := tx.CreateBucketIfNotExists([]byte(BUCKET_CREDENTIALS)); err != nil {
   return err
  }
  return nil
 })
 
 if err != nil {
  return nil, err
 }
 return &BoltDB{conn: db}, nil
}

func (db *BoltDB) Close() error {
 return db.conn.Close()
}

func (db *BoltDB) SaveSMTPConfig(cfg domain.SMTPConfig) error {
 return db.conn.Update(func(tx *bolt.Tx) error {
  bucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_NAME))
  if err != nil {
   return err
  }
  data, _ := json.Marshal(config)
  return bucket.Put([]byte(SMTP_KEY),data)
 })
}

func (db *BoltDB) GetSMTPConfig() (domain.SMTPConfig, error) {
 var config domain.SMTPConfig
 err := db.conn.View(func(tx *bolt.Tx) error {
 bucket := tx.Bucket([]byte(BUCKET_NAME))
 if bucket == nil {
  return fmt.Errorf("Config Bucket not found.")
 }
 data := bucket.Get([]byte(config))
 if data == nil {
  return fmt.Errorf("no SMTP config found")
 } 
 return json.Unmarshal(v, &config)
 })
 return config, err
}

func (db *BoltDB) AddRecipient(email string) error {
 return db.conn.Update(func(tx *bbolt.TX) error {
  bucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_RECEPIENTS))
  if err != nil {
   return err
  }
  return bucket.Put([]byte(email), []byte(email))
 })
}

func (db *BoltDB) RemoveRecipient(email string) error {
 return db.conn.Update(func(tx *bbolt.TX) error {
  bucket := tx.Bucket([]byte(BUCKET_RECEPIENTS))
  if bucket == nil {
   return nil
  }
  return bucket.Delete([]byte(email))
 })
}

func (db *BoltDB) ListRecipient() ([]string, error) {
 var emails []string
 err := db.conn.View(func(tx *bbolt.TX) error {
  bucket := tx.Bucket([]byte(BUCKET_RECEPIENTS))
   if bucket == nil {
    return nil
   }

  return bucket.Foreach(func(k, v []byte) error {
   emails = append(emails, string(v))
   return nil
  })
 })
 return emails, err
}

func (db *BoltDB) SavePassword(host string, password string, masterKey []byte) error {
 encrypted, err := encrypt([]byte(password),masterKey)
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
   return fmt.Errorf("Password not found for host.: %s",host)
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
 vKey  := b.Get([]byte(SSH_KEY_PATH_KEY))
  if vUser == nil || vKey == nil {
   return errors.New("SSH configuration not found.")
  } 

 user = string(vUser)
 keyPath = string(vKey)
 return nil
})
return user, keyPath, err
}

