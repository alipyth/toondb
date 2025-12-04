package db

import (
        "fmt"
        "log"
        "strings"

        "github.com/dgraph-io/badger/v3"
)

type Database struct {
        db *badger.DB
}

type Record struct {
        Collection string `json:"collection"`
        Key        string `json:"key"`
        Data       string `json:"data"`
}

func NewDatabase(path string) (*Database, error) {
        opts := badger.DefaultOptions(path)
        opts.Logger = nil // Disable badger logging
        
        db, err := badger.Open(opts)
        if err != nil {
                return nil, fmt.Errorf("failed to open badger database: %w", err)
        }

        return &Database{db: db}, nil
}

func (d *Database) Close() error {
        return d.db.Close()
}

func (d *Database) Get(collection, key string) (string, error) {
        var data string
        err := d.db.View(func(txn *badger.Txn) error {
                item, err := txn.Get([]byte(fmt.Sprintf("%s:%s", collection, key)))
                if err != nil {
                        return err
                }
                
                return item.Value(func(val []byte) error {
                        data = string(val)
                        return nil
                })
        })

        if err == badger.ErrKeyNotFound {
                return "", fmt.Errorf("key not found")
        }

        return data, err
}

func (d *Database) Set(collection, key, data string) error {
        return d.db.Update(func(txn *badger.Txn) error {
                return txn.Set([]byte(fmt.Sprintf("%s:%s", collection, key)), []byte(data))
        })
}

func (d *Database) Delete(collection, key string) error {
        return d.db.Update(func(txn *badger.Txn) error {
                return txn.Delete([]byte(fmt.Sprintf("%s:%s", collection, key)))
        })
}

func (d *Database) DeleteCollection(collection string) error {
        return d.db.Update(func(txn *badger.Txn) error {
                it := txn.NewIterator(badger.DefaultIteratorOptions)
                defer it.Close()
                
                prefix := []byte(collection + ":")
                
                for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
                        item := it.Item()
                        key := item.Key()
                        if err := txn.Delete(key); err != nil {
                                return err
                        }
                }
                return nil
        })
}

func (d *Database) GetCollectionKeys(collection string) ([]string, error) {
        var keys []string
        
        err := d.db.View(func(txn *badger.Txn) error {
                it := txn.NewIterator(badger.DefaultIteratorOptions)
                defer it.Close()
                
                prefix := []byte(collection + ":")
                
                for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
                        item := it.Item()
                        key := string(item.Key())
                        
                        // Remove the collection prefix to get the actual key
                        if strings.HasPrefix(key, collection+":") {
                                actualKey := key[len(collection)+1:]
                                keys = append(keys, actualKey)
                        }
                }
                return nil
        })
        
        return keys, err
}

func (d *Database) GetCollections() (map[string][]string, error) {
        collections := make(map[string][]string)
        
        err := d.db.View(func(txn *badger.Txn) error {
                it := txn.NewIterator(badger.DefaultIteratorOptions)
                defer it.Close()
                
                for it.Rewind(); it.Valid(); it.Next() {
                        item := it.Item()
                        key := string(item.Key())
                        
                        parts := strings.Split(key, ":")
                        if len(parts) == 2 {
                                collection := parts[0]
                                keyName := parts[1]
                                collections[collection] = append(collections[collection], keyName)
                        }
                }
                return nil
        })
        
        return collections, err
}

func (d *Database) Backup() ([]Record, error) {
        var records []Record
        
        err := d.db.View(func(txn *badger.Txn) error {
                it := txn.NewIterator(badger.DefaultIteratorOptions)
                defer it.Close()
                
                for it.Rewind(); it.Valid(); it.Next() {
                        item := it.Item()
                        key := string(item.Key())
                        
                        parts := strings.Split(key, ":")
                        if len(parts) == 2 {
                                collection := parts[0]
                                keyName := parts[1]
                                
                                err := item.Value(func(val []byte) error {
                                        record := Record{
                                                Collection: collection,
                                                Key:        keyName,
                                                Data:       string(val),
                                        }
                                        records = append(records, record)
                                        return nil
                                })
                                
                                if err != nil {
                                        return err
                                }
                        }
                }
                return nil
        })
        
        return records, err
}

func (d *Database) Restore(records []Record) error {
        return d.db.Update(func(txn *badger.Txn) error {
                for _, record := range records {
                        key := fmt.Sprintf("%s:%s", record.Collection, record.Key)
                        err := txn.Set([]byte(key), []byte(record.Data))
                        if err != nil {
                                log.Printf("Failed to restore record %s:%s: %v", record.Collection, record.Key, err)
                                return err
                        }
                }
                return nil
        })
}