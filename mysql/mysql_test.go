package mysql

import "testing"

// - mysql://tanya:010203@localhost:3306/PROJECT (success)
// - mysql://tanya:010203@localhost:3306/
// - mysql://tanya:010203@/PROJECT
// - mysql://tanya@localhost:3306/PROJECT(success)
// - mysql://tanya@local/PROJECT
// - mysql://010203@localhost:3306/PROJECT
// - mysql://localhost:3306/PROJECT
// - mysql://tanya:010203@localhost:3306/PROJECT/?a=3
// - mysql://tanya:010203@localhost:3306/PROJECT/#a
// - mysql://
// - test://
func TestConnect(t *testing.T) {

}

// - mysql://tanya@localhost:3306/PROJECT, TABLE created
// - mysql://tanya@localhost:3306/PROJECT, Connection drop, TABLE not created, error
func TestInitialize(t *testing.T) {

}

//- Protocol registered, 'SELECT COUNT (*) FROM Narada4D', true
//- Protocol not registered, 'SELECT COUNT (*) FROM Narada4D', false
//- Protocol registered, 'SELECT COUNT (*) FROM Narada4D', connection
//drop, false (reconnected automaticaly - true)???
func TestInitialized(t *testing.T) {

}

// - SH, SH, UN, UN
// - SH, EX(block), UN(SH), EX, UN(EX)
// - EX, SH(block), UN(EX), SH, UN(SH)
// - SH, EX(block), SH(block), UN(SH), EX, UN(EX), SH, UN(SH)
func TestShExLock(t *testing.T) {

}

// - UN, error
func TestUnlock(t *testing.T) {

}

// - Protocol registered, 'SELECT val FROM Narada4D WHERE var=`version`' (success)
// - Protocol not registered, 'SELECT val FROM Narada4D WHERE var=`version`', panic
func TestGet(t *testing.T) {

}

// - Protocol registered, sqlSetVersion, val=43, success
// - Protocol registered, sqlSetVersion, val=43.0, success
// - Protocol registered, sqlSetVersion, val=43.0.1, success
// - Protocol registered, sqlSetVersion, val="", panic
// - Protocol registered, sqlSetVersion, val=0, ?
// - Protocol registered, sqlSetVersion, val=-18, panic
// - Protocol registered, sqlSetVersion, val=rat, panic
// - Protocol not registered, sqlSetVersion, val=43, panic
func TestSet(t *testing.T) {

}
