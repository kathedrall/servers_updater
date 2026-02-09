package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
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

func (db *BoltDB) GetMachineByHost(host string) (domain.Machine, error) {
	var machine domain.Machine

	err := db.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_MACHINES))
		if bucket == nil {
			return fmt.Errorf("machines bucket not found")
		}

		data := bucket.Get([]byte(host))
		if data == nil {
			return fmt.Errorf("machine with host %s not found", host)
		}

		if err := json.Unmarshal(data, &machine); err != nil {
			return fmt.Errorf("error unmarshaling machine data: %v", err)
		}

		return nil
	})

	if err != nil {
		return machine, err
	}

	// Descriptografar senha APÓS a transação se existir e estiver criptografada
	if machine.Password != "" {
		isEncrypted := isEncryptedData([]byte(machine.Password))
		
		if isEncrypted {
			masterKey, err := db.getMasterKey()
			if err != nil {
				return machine, fmt.Errorf("erro ao obter chave master: %v", err)
			}
			
			// Decodificar de base64
			encryptedData, err := base64.StdEncoding.DecodeString(machine.Password)
			if err != nil {
				// Se falhar ao decodificar, pode ser senha antiga em texto plano
				return machine, nil
			}
			
			decryptedPassword, err := decrypt(encryptedData, masterKey)
			if err != nil {
				// Se falhar ao descriptografar, pode ser senha antiga em texto plano
				// Mantém a senha como está (compatibilidade com dados antigos)
				return machine, nil
			}
			machine.Password = string(decryptedPassword)
		}
	}

	return machine, nil
}

// isEncryptedData verifica se os dados parecem estar criptografados
// Agora verifica se é uma string base64 válida
func isEncryptedData(data []byte) bool {
	str := string(data)
	// Se tem menos de 16 chars (12 bytes nonce + 16 bytes tag codificados), provavelmente não está criptografado
	if len(str) < 16 {
		return false
	}
	
	// Tentar decodificar como base64
	_, err := base64.StdEncoding.DecodeString(str)
	return err == nil
}

func (db *BoltDB) SaveMachine(m domain.Machine) error {
	// Criptografar senha ANTES da transação se necessário
	var machineToSave = m
	if m.Password != "" {
		masterKey, err := db.getMasterKey()
		if err != nil {
			return fmt.Errorf("erro ao obter chave master: %v", err)
		}
		encryptedPassword, err := encrypt([]byte(m.Password), masterKey)
		if err != nil {
			return fmt.Errorf("erro ao criptografar senha: %v", err)
		}
		// Codificar em base64 para evitar corrupção durante JSON
		machineToSave.Password = base64.StdEncoding.EncodeToString(encryptedPassword)
	}

	return db.conn.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_MACHINES))
		if err != nil {
			return err
		}

		// SEMPRE usar o Host como chave para consistência
		key := machineToSave.Host

		data, err := json.Marshal(machineToSave)
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

// getMasterKey obtém ou gera uma chave master para criptografia
func (db *BoltDB) getMasterKey() ([]byte, error) {
	var masterKey []byte
	err := db.conn.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_SETTINGS))
		if b == nil {
			return fmt.Errorf("settings bucket not found")
		}

		// Tentar buscar chave existente
		existingKey := b.Get([]byte("master_key"))
		if existingKey != nil {
			masterKey = make([]byte, len(existingKey))
			copy(masterKey, existingKey)
			return nil
		}

		return fmt.Errorf("master key not found")
	})

	if err != nil {
		// Chave não existe, criar uma nova
		return db.createMasterKey()
	}

	return masterKey, nil
}

// createMasterKey cria uma nova chave master
func (db *BoltDB) createMasterKey() ([]byte, error) {
	var masterKey []byte
	err := db.conn.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(BUCKET_SETTINGS))
		if err != nil {
			return err
		}

		// Verificar novamente se a chave foi criada entre as chamadas
		existingKey := b.Get([]byte("master_key"))
		if existingKey != nil {
			masterKey = make([]byte, len(existingKey))
			copy(masterKey, existingKey)
			return nil
		}

		// Gerar nova chave de 32 bytes (AES-256)
		newKey := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
			return err
		}

		// Salvar a nova chave
		if err := b.Put([]byte("master_key"), newKey); err != nil {
			return err
		}

		masterKey = make([]byte, 32)
		copy(masterKey, newKey)
		return nil
	})
	return masterKey, err
}

// HasStoredCredentials verifica se uma máquina possui credenciais armazenadas
func (db *BoltDB) HasStoredCredentials(host string) bool {
	err := db.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_MACHINES))
		if bucket == nil {
			return fmt.Errorf("machines bucket not found")
		}

		data := bucket.Get([]byte(host))
		if data == nil {
			return fmt.Errorf("machine not found")
		}

		var machine domain.Machine
		if err := json.Unmarshal(data, &machine); err != nil {
			return fmt.Errorf("error unmarshaling machine data: %v", err)
		}

		if machine.Password == "" && machine.KeyPath == "" {
			return fmt.Errorf("no credentials stored")
		}

		return nil
	})
	return err == nil
}

// UpdateMachineCredentials atualiza apenas as credenciais de uma máquina sem afetar outras informações
func (db *BoltDB) UpdateMachineCredentials(host, password, keyPath string) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_MACHINES))
		if bucket == nil {
			return fmt.Errorf("machines bucket not found")
		}

		data := bucket.Get([]byte(host))
		if data == nil {
			return fmt.Errorf("machine with host %s not found", host)
		}

		var machine domain.Machine
		if err := json.Unmarshal(data, &machine); err != nil {
			return fmt.Errorf("error unmarshaling machine data: %v", err)
		}

		// Atualizar apenas as credenciais
		machine.Password = password
		machine.KeyPath = keyPath

		// Salvar de volta
		updatedData, err := json.Marshal(machine)
		if err != nil {
			return fmt.Errorf("error marshaling updated machine data: %v", err)
		}

		return bucket.Put([]byte(host), updatedData)
	})
}

// GetMachineCredentials retorna apenas as credenciais de uma máquina
func (db *BoltDB) GetMachineCredentials(host string) (password, keyPath string, err error) {
	err = db.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_MACHINES))
		if bucket == nil {
			return fmt.Errorf("machines bucket not found")
		}

		data := bucket.Get([]byte(host))
		if data == nil {
			return fmt.Errorf("machine with host %s not found", host)
		}

		var machine domain.Machine
		if err := json.Unmarshal(data, &machine); err != nil {
			return fmt.Errorf("error unmarshaling machine data: %v", err)
		}

		password = machine.Password
		keyPath = machine.KeyPath
		return nil
	})
	return password, keyPath, err
}

// ClearMachinePassword remove a senha armazenada de uma máquina (útil quando a autenticação falha)
func (db *BoltDB) ClearMachinePassword(host string) error {
	return db.conn.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_MACHINES))
		if bucket == nil {
			return fmt.Errorf("machines bucket not found")
		}

		data := bucket.Get([]byte(host))
		if data == nil {
			return fmt.Errorf("machine with host %s not found", host)
		}

		var machine domain.Machine
		if err := json.Unmarshal(data, &machine); err != nil {
			return fmt.Errorf("error unmarshaling machine data: %v", err)
		}

		// Limpar apenas a senha
		machine.Password = ""

		// Salvar de volta
		updatedData, err := json.Marshal(machine)
		if err != nil {
			return fmt.Errorf("error marshaling updated machine data: %v", err)
		}

		return bucket.Put([]byte(host), updatedData)
	})
}
