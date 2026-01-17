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
 bucketSettings = "Settings"
 bucketCredentials = "Credentials"
 smtpKey = "smtp_config"
 poolLimitKey = "pool_limit"
 poolLimitDefault = 10

 sshUserKey = "ssh_user"
 sshKeyPathKey = "ssh_key_path"
)

func InitDB(path string) (*BoltDB, error) {
 db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second}) 
 if err != nil {
  return nil, fmt.Errorf("error when starting bbolt: %w", err)
 }

 err = db.Update(func(tx *bbolt.Tx) error {
  if _, err := tx.CreateBucketIfNotExists([]byte(bucketSettings)); err != nil {
   return err
 }
  if _, err := tx.CreateBucketIfNotExists([]byte(bucketCredentials)); err != nil {
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
 return db.conn.Update(func(tx *bbolt.Tx) error {
  b := tx.Bucket([]byte(bucketSettings))
  data, err := json.Marshal(cfg)
   if err != nil {
    return err
   }
  return b.Put([]byte(smtpKey),data)
 })
}

func (db *BoltDB) GetSMTPConfig() (domain.SMTPConfig, error) {
 var cfg domain.SMTPConfig
 err := db.conn.View(func(tx *bbolt.Tx) error {
  b := tx.Bucket([]byte(bucketSettings))
  v := b.Get([]byte(smtpKey))
   if v == nil {
    return errors.New("SMTP configuration not found.")
   }
   return json.Unmarshal(v, &cfg)
 })
 return cfg, err
}

func (db *BoltDB) SavePassword(host string, password string, masterKey []byte) error {
 encrypted, err := encrypt([]byte(password),masterKey)
 if err != nil {
  return err
 }
 return db.conn.Update(func(tx *bbolt.Tx) error {
  b := tx.Bucket([]byte(bucketCredentials))
  return b.Put([]byte(host), encrypted)
 })
}

func (db *BoltDB) GetPassword(host string, masterKey []byte) (string, error) {
 var decrypted []byte
 err := db.conn.View(func(tx *bbolt.Tx) error {
  b := tx.Bucket([]byte(bucketCredentials))
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
  b := tx.Bucket([]byte(bucketSettings))
  val := fmt.Sprintf("%d", limit)
   return b.Put([]byte(poolLimitKey), []byte(val))
 })
}

func (db *BoltDB) GetPoolLimit() (int, error) {
 var limit int
 err := db.conn.View(func(tx *bbolt.Tx) error {
  b := tx.Bucket([]byte(bucketSettings))
  v := b.Get([]byte(poolLimitKey))
   if v == nil {
    limit = poolLimitDefault
    return nil
   } 
  _, err := fmt.Sscanf(string(v), "%d", &limit)
    return err
 })
 return limit, err
}

func (db *BoltDB) SaveSSHConfig(user string, keyPath string) error {
 return db.conn.Update(func(tx *bbolt.Tx) error {
  b := tx.Bucket([]byte(bucketSettings))
  if b == nil {
   return fmt.Errorf("bucket %s not found", bucketSettings)
  }

  if err := b.Put([]byte(sshUserKey), []byte(user)); err != nil {
   return err
  }
  return b.Put([]byte(sshKeyPathKey), []byte(keyPath)) 
 })
}

func (db *BoltDB) GetSSHConfig() (string, string, error) {
 var user, keyPath string

 err := db.conn.View(func(tx *bbolt.Tx) error {
  b := tx.Bucket([]byte(bucketSettings))
  if b == nil {
   return fmt.Errorf("bucket %s not found", bucketSettings)
  }

 vUser := b.Get([]byte(sshUserKey))
 vKey  := b.Get([]byte(sshKeyPathKey))
  if vUser == nil || vKey == nil {
   return errors.New("SSH configuration not found.")
  } 

 user = string(vUser)
 keyPath = string(vKey)
 return nil
})
return user, keyPath, err
}

