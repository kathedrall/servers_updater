package security

 import "testing"
 

 func TestCheckOSV(t *testing.T) {

 t.Run("Detect known vulnerability", func(t *testing.T) {	 
  isVun, info := CheckOSV("openssl", "1.1.1f-ubuntu2")
  if !isVun {
   t.Error("It should have detected a vulnerability in OpenSSL 1.1.1f")
  }
  if info == "" {
   t.Error("The information field should not be empty.")
  } 
 })

 t.Run("Secure package or nonexistent package", func(t *testing.T) {
  isVun, _ := CheckOSV("Servers Update dummy package", "9.2.4") 

  if isVun {
   t.Error("It should not detect a vulnerability in a non-existent packet.")
  }
 })
}

 
