package db

import (
	"encoding/json"
	"os"
	"servers_updater/internal/domain"
	"strings"
	"testing"

	"go.etcd.io/bbolt"
)

func setupTestDB(t *testing.T) (*BoltDB, string) {
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	path := "test_" + safeName + ".db"
	os.Remove(path)
	database, err := InitDB(path)
	if err != nil {
		t.Fatalf("Error starting database for %s: %v", t.Name(), err)
	}
	return database, path
}

func TestSaveAndGetSMTPConfig(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	config := domain.SMTPConfig{
		Host: "smtp.test.com",
		User: "test@user.com",
	}
	if err := db.SaveSMTPConfig(config); err != nil {
		t.Errorf("It shouldn't return an error when saving the SMTP file.: %v", err)
	}

	saved, err := db.GetSMTPConfig()
	if err != nil || saved.Host != config.Host {
		t.Errorf("Error retrieving SMTP configuration.: %v", err)
	}
}

func TestSSHConfigPersistence(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	t.Run("Save and Retrieve", func(t *testing.T) {
		expectedUser := "root"
		expectedKey := "/home/user/.ssh/id_rsa_test"

		err := db.SaveSSHConfig(expectedUser, expectedKey)
		if err != nil {
			t.Fatalf("Failed to save SSH config: %v", err)
		}

		user, key, err := db.GetSSHConfig()
		if err != nil {
			t.Fatalf("Failed to retrieve SSH config: %v", err)
		}

		if user != expectedUser {
			t.Errorf("Expected user %s, got %s", expectedUser, user)
		}
		if key != expectedKey {
			t.Errorf("Expected key path %s, got %s", expectedKey, key)
		}
	})

	t.Run("Error when not found", func(t *testing.T) {
		dbClean, pathClean := setupTestDB(t)
		defer func() { dbClean.Close(); os.Remove(pathClean) }()

		_, _, err := dbClean.GetSSHConfig()
		if err == nil {
			t.Error("Expected error when SSH config is missing, got nil")
		}
	})
}

func TestPasswordEncryption(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	masterKey := []byte("key-must-have-exactly-32-bytes!!")
	host, pass := "srv-01", "my-secret-pass"

	t.Run("decrypted successfully", func(t *testing.T) {
		db.SavePassword(host, pass, masterKey)
		decrypted, err := db.GetPassword(host, masterKey)
		if err != nil || decrypted != pass {
			t.Errorf("Expected pass %s, received %s (err: %v)", pass, decrypted, err)
		}
	})

	t.Run("Incorrect key error", func(t *testing.T) {
		wrongKey := []byte("wrong-key-must-have-32-bytes-too")
		_, err := db.GetPassword(host, wrongKey)

		if err == nil {
			t.Error("It should have failed with the incorrect master key, but it succeeded.")
		}
	})
}

func TestPoolLimitPersistence(t *testing.T) {
	t.Run("Save and recover limit", func(t *testing.T) {
		db, path := setupTestDB(t)
		defer func() { db.Close(); os.Remove(path) }()

		limit := 42
		db.SavePoolLimit(limit)
		val, _ := db.GetPoolLimit()
		if val != limit {
			t.Errorf("Expected %d, received %d", limit, val)
		}
	})

	t.Run("Default value when empty", func(t *testing.T) {
		db, path := setupTestDB(t)
		defer func() { db.Close(); os.Remove(path) }()

		val, _ := db.GetPoolLimit()
		if val != 10 {
			t.Errorf("The default return value should be 10. But it received %d", val)
		}
	})
}

func TestMachineCRUD(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	userPtr := "admin"
	portPtr := 22
	machine := domain.Machine{
		Host:       "192.168.1.50",
		User:       userPtr,
		Port:       portPtr,
		PrettyName: "Ubuntu 22.04 LTS",
		Status:     "ONLINE",
	}

	if err := db.SaveMachine(machine); err != nil {
		t.Fatalf("Error save machine: %v", err)
	}

	machines, err := db.GetAllMachines()
	if err != nil {
		t.Fatalf("Error reading the machines: %v", err)
	}

	if len(machines) != 1 {
		t.Errorf("waiting for 1 machine, bug return %d:", len(machines))
	}

	if machines[0].Host != machine.Host {
		t.Errorf("Corrupted data. I was expecting host %s, but received %s", machine.Host, machines[0].Host)
	}
}

func TestMachinePasswordEncryption(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	// Máquina de teste com senha
	testMachine := domain.Machine{
		Host:       "test.crypto.com",
		User:       "testuser",
		Password:   "minhasenhasecreta123",
		OSName:     "ubuntu",
		OSVersion:  "22.04",
		PrettyName: "Ubuntu 22.04 LTS",
	}

	t.Run("SaveAndRetrieveEncryptedPassword", func(t *testing.T) {
		// Salvar máquina com senha
		err := db.SaveMachine(testMachine)
		if err != nil {
			t.Fatalf("Erro ao salvar máquina: %v", err)
		}

		// Recuperar máquina
		retrievedMachine, err := db.GetMachineByHost(testMachine.Host)
		if err != nil {
			t.Fatalf("Erro ao recuperar máquina: %v", err)
		}

		// Verificar se a senha foi descriptografada corretamente
		if retrievedMachine.Password != testMachine.Password {
			t.Errorf("Senha não foi descriptografada corretamente. Esperado: %s, Obtido: %s",
				testMachine.Password, retrievedMachine.Password)
		}

		// Verificar se outros dados foram preservados
		if retrievedMachine.Host != testMachine.Host {
			t.Errorf("Host não foi preservado. Esperado: %s, Obtido: %s",
				testMachine.Host, retrievedMachine.Host)
		}

		if retrievedMachine.User != testMachine.User {
			t.Errorf("User não foi preservado. Esperado: %s, Obtido: %s",
				testMachine.User, retrievedMachine.User)
		}
	})

	t.Run("PasswordIsEncryptedInDatabase", func(t *testing.T) {
		// Verificar diretamente no banco se a senha está criptografada
		var storedPassword string
		err := db.conn.View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(BUCKET_MACHINES))
			if bucket == nil {
				t.Fatal("Bucket MACHINES não encontrado")
			}

			data := bucket.Get([]byte(testMachine.Host))
			if data == nil {
				t.Fatal("Máquina não encontrada no banco")
			}

			var storedMachine domain.Machine
			if err := json.Unmarshal(data, &storedMachine); err != nil {
				return err
			}

			storedPassword = storedMachine.Password
			return nil
		})

		if err != nil {
			t.Fatalf("Erro ao ler dados criptografados: %v", err)
		}

		// A senha armazenada deve ser diferente da senha original (criptografada)
		if storedPassword == testMachine.Password {
			t.Error("Senha está sendo salva em texto plano! Deve estar criptografada.")
		}

		// A senha criptografada não deve estar vazia
		if storedPassword == "" {
			t.Error("Senha criptografada está vazia")
		}

		t.Logf("Senha original: %s", testMachine.Password)
		t.Logf("Senha criptografada (primeiros 20 chars): %s...", 
			func() string {
				if len(storedPassword) > 20 {
					return storedPassword[:20]
				}
				return storedPassword
			}())
	})

	t.Run("EmptyPasswordHandling", func(t *testing.T) {
		// Testar máquina sem senha
		machineNoPassword := domain.Machine{
			Host:       "no-password.test.com",
			User:       "testuser",
			OSName:     "debian",
		}

		err := db.SaveMachine(machineNoPassword)
		if err != nil {
			t.Fatalf("Erro ao salvar máquina sem senha: %v", err)
		}

		retrieved, err := db.GetMachineByHost(machineNoPassword.Host)
		if err != nil {
			t.Fatalf("Erro ao recuperar máquina sem senha: %v", err)
		}

		if retrieved.Password != "" {
			t.Errorf("Máquina sem senha deveria ter Password vazio, mas tem: %s", retrieved.Password)
		}
	})
}

func TestMasterKeyGeneration(t *testing.T) {
	db, path := setupTestDB(t)
	defer func() { db.Close(); os.Remove(path) }()

	t.Run("MasterKeyIsGenerated", func(t *testing.T) {
		key1, err := db.getMasterKey()
		if err != nil {
			t.Fatalf("Erro ao obter chave master: %v", err)
		}

		if len(key1) != 32 {
			t.Errorf("Chave master deveria ter 32 bytes, mas tem %d", len(key1))
		}

		// Segunda chamada deve retornar a mesma chave
		key2, err := db.getMasterKey()
		if err != nil {
			t.Fatalf("Erro ao obter chave master novamente: %v", err)
		}

		if string(key1) != string(key2) {
			t.Error("Chave master deveria ser a mesma em chamadas subsequentes")
		}
	})
}

func TestEncryptionFunctions(t *testing.T) {
	testData := "dados para testar criptografia"
	key := make([]byte, 32) // AES-256 key
	copy(key, []byte("esta-e-uma-chave-de-32-bytes-ok"))

	t.Run("EncryptAndDecrypt", func(t *testing.T) {
		// Criptografar
		encrypted, err := encrypt([]byte(testData), key)
		if err != nil {
			t.Fatalf("Erro ao criptografar: %v", err)
		}

		// Dados criptografados devem ser diferentes dos originais
		if string(encrypted) == testData {
			t.Error("Dados criptografados são iguais aos originais")
		}

		// Descriptografar
		decrypted, err := decrypt(encrypted, key)
		if err != nil {
			t.Fatalf("Erro ao descriptografar: %v", err)
		}

		// Dados descriptografados devem ser iguais aos originais
		if string(decrypted) != testData {
			t.Errorf("Descriptografia falhou. Esperado: %s, Obtido: %s", testData, string(decrypted))
		}
	})

	t.Run("DecryptWithWrongKey", func(t *testing.T) {
		// Criptografar com uma chave
		encrypted, err := encrypt([]byte(testData), key)
		if err != nil {
			t.Fatalf("Erro ao criptografar: %v", err)
		}

		// Tentar descriptografar com chave errada
		wrongKey := make([]byte, 32)
		copy(wrongKey, []byte("esta-e-uma-chave-diferente-32b"))

		_, err = decrypt(encrypted, wrongKey)
		if err == nil {
			t.Error("Deveria falhar ao descriptografar com chave errada")
		}
	})
}
