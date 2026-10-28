package main

import "errors"

var ErrNoTransaction = errors.New("no transaction")

type KVStore struct {
	// TODO
}

func NewKVStore() *KVStore { panic("TODO") }

func (s *KVStore) Set(key, value string)         { panic("TODO") }
func (s *KVStore) Get(key string) (string, bool) { panic("TODO") }
func (s *KVStore) Delete(key string)             { panic("TODO") }
func (s *KVStore) Count(value string) int        { panic("TODO") }
func (s *KVStore) Begin()                        { panic("TODO") }
func (s *KVStore) Rollback() error               { panic("TODO") }
func (s *KVStore) Commit() error                 { panic("TODO") }
