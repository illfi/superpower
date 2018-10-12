package api

import (
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
	"ill.fi/superpower/server/lib"
	"log"
	"os"
	"strconv"
)

func CreateDBHandler(dblu, dblf, dbli string) *DBHandler {
	opts := badger.DefaultOptions
	opts.Dir = dblu
	opts.ValueDir = dblu
	opts.ValueLogLoadingMode = options.FileIO
	dbu, err := badger.Open(opts)
	lib.Check(err, func(e error) {
		log.Fatalf("Failed to open database, " + e.Error())
		os.Exit(1)
	})

	opts = badger.DefaultOptions
	opts.Dir = dblf
	opts.ValueDir = dblf
	opts.ValueLogLoadingMode = options.FileIO
	dbf, err := badger.Open(opts)
	lib.Check(err, func(e error) {
		log.Fatalf("Failed to open database, " + e.Error())
		os.Exit(1)
	})

	opts = badger.DefaultOptions
	opts.Dir = dbli
	opts.ValueDir = dbli
	opts.ValueLogLoadingMode = options.FileIO
	dbi, err := badger.Open(opts)
	lib.Check(err, func(e error) {
		log.Fatalf("Failed to open database, " + e.Error())
		os.Exit(1)
	})

	h := &DBHandler{
		Users:   dbu,
		Files:   dbf,
		Invites: dbi,
	}
	return h
}

// this file creates queries and such
// TODO: rewrite this using actual queries instead of querying the accounts.

//////////////
// accounts //
//////////////

func GetAccountsByDisplayName(name string, db *badger.DB) []Account {
	a := GetAccounts(db)
	var r []Account
	for _, x := range a {
		if x.DisplayName == name {
			r = append(r, x)
		}
	}
	return r
}

// i should really use reflections for this tbqh
func GetAccountByID(id int, db *badger.DB) *Account {
	a := GetAccounts(db)
	for _, x := range a {
		if x.ID == id {
			return &x
		}
	}
	return nil
}

func GetParentAccount(acc Account, db *badger.DB) *Account {
	a := GetAccounts(db)
	for _, x := range a {
		for _, i := range x.Invites {
			if i == acc.ID {
				return &x
			}
		}
	}
	return nil
}

func GetAccountByEmail(email string, db *badger.DB) *Account {
	a := GetAccounts(db)
	for _, x := range a {
		if x.Email == email {
			return &x
		}
	}
	return nil
}

func AddAccount(a Account, db *badger.DB) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(strconv.Itoa(a.ID)), []byte(SerializeAccount(a)))
	})
}

func RemoveAccount(a Account, db *badger.DB) error {
	return db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			i := it.Item()
			k := string(i.Key())
			id, e := strconv.Atoi(k)
			lib.NOPCheck(e)
			if id == a.ID {
				txn.Delete(i.Key())
			}
		}
		return nil
	})
}

func GetLastID(db *badger.DB) int {
	a := GetLastAccount(db)
	if a == nil {
		return 0
	} else {
		return a.ID + 1
	}
}

func GetLastAccount(db *badger.DB) *Account {
	a := GetAccounts(db)
	highest := 0
	for _, x := range a { // cant guarantee order..?
		if x.ID > highest {
			highest = x.ID
		}
	}
	return GetAccountByID(highest, db)
}

func GetAccounts(db *badger.DB) []Account {
	var accounts []Account
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			i := it.Item()
			k := string(i.Key())
			v, e := i.Value()
			lib.NOPCheck(e)
			id, e := strconv.Atoi(k)
			lib.NOPCheck(e)
			dat := string(v)
			accounts = append(accounts, ParseAccount(Account{ID: id}, dat))
		}
		return nil
	})
	lib.NOPCheck(err)
	return accounts
}

/////////////
// invites //
/////////////

func AddInvite(i Invite, db *badger.DB) {
	db.Update(func(txn *badger.Txn) error {
		txn.Set([]byte(strconv.Itoa(i.From)), []byte(string(i.AuthorizedEmail)))
		return nil
	})
}

func IsInvited(email string, db *badger.DB) bool {
	invs := GetInvites(db)
	for _, i := range invs {
		if i.AuthorizedEmail == email {
			return true
		}
	}
	return false
}

func GetInvites(db *badger.DB) []Invite {
	var invites []Invite
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			i := it.Item()
			k := string(i.Key())
			v, e := i.Value()
			lib.NOPCheck(e)
			id, e := strconv.Atoi(k)
			lib.NOPCheck(e)
			invites = append(invites, Invite{
				AuthorizedEmail: string(v),
				From:            id,
			})
		}
		return nil
	})
	lib.NOPCheck(err)
	return invites
}
